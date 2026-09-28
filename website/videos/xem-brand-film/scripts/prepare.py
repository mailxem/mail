"""Freeze source-derived effect code and prepare the authoritative Xem mark.

Run with uv run --with pillow scripts/prepare.py. No website sources are modified.
The registry effect mathematics/shader are retained; the canvas adapter lets the
same effect participate in a continuously moving, clipped montage camera.
"""
from pathlib import Path
import re
from PIL import Image

ROOT = Path(__file__).resolve().parents[1]
im = Image.open(ROOT / 'assets/brand/xem-mark.png').convert('RGB')
alpha = Image.new('RGBA', im.size)
pixels = []
for r, g, b in im.getdata():
    a = max(0, min(255, round((240 - max(r, g, b)) * 255 / 225)))
    pixels.append((34, 37, 31, a))
alpha.putdata(pixels)
alpha = alpha.crop(alpha.getbbox())
alpha.save(ROOT / 'assets/brand/xem-mark-alpha.png')

carousel = (ROOT / 'compositions/carousel-circle-4.html').read_text()
scripts = re.findall(r'<script>(.*?)</script>', carousel, re.S)
rig = scripts[0]
data = scripts[1].split('(async function')[0].replace('const DATA =', 'window.XEM_CAROUSEL_DATA =')
(ROOT / 'assets/vendor/carousel-rig.js').write_text('// Adapted from installed HyperFrames Circle Carousel 4.\n' + rig + '\n' + data)

halftone = (ROOT / 'compositions/halftone-field.html').read_text()
code = re.findall(r'<script>(.*?)</script>', halftone, re.S)[0]
code = code.replace('(function () {', 'window.createXemHalftone = function (canvas) {', 1)
a = code.index('        var V =')
b = code.index('        // #RGB', a)
code = code[:a] + '''        var settings = {frequency: 2, speed: 1, cellSize: 14, gamma: 10, paletteBias: 1.5,
          palette1: '#5b3cc4', palette2: '#e7d8fa', palette3: '#edf09b', palette4: '#5b3cc4', background: '#22251f'};
        function raw(id) { return settings[id]; }
        function num(id, fallback) { return settings[id] == null ? fallback : settings[id]; }
        function bool(id) { return false; }
''' + code[b:]
a = code.index('        document.getElementById("hlf-backdrop")')
b = code.index('        var gl =', a)
code = code[:a] + code[b:]
a = code.index('        window.__timelines =')
code = code[:a] + '''        draw(0);
        return {draw: draw, canvas: canvas, ready: ready};
      };
'''
(ROOT / 'assets/vendor/halftone-field.js').write_text('// Shader and uniform mapping from installed HyperFrames Halftone Field.\n'+code)
print('Prepared transparent mark and two registry effect adapters.')
