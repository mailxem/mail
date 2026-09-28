"""Build the Studio timeline from the approved frame map."""
from pathlib import Path
import json
from html import escape
P=Path(__file__).resolve().parents[1]
plan=json.loads((P/'scene-plan.json').read_text())
scenes=plan['scenes']
faces='\n'.join(f'@font-face {{font-family:"{family}";font-weight:{weight};src:url("assets/fonts/{slug}-{weight}.ttf") format("truetype");font-display:block;}}' for family,slug,weights in [('DM Sans','dm-sans',[400,500,600,700]),('EB Garamond','eb-garamond',[400,500])] for weight in weights)
def track(s):
    n=int(s['id'])
    return next(i for i,end in enumerate([6,10,13,19,22,25,28]) if n<=end)
clips='\n'.join(f'''<canvas id="canvas-{s['id']}" class="clip" data-start="{s['start']/30:.12f}" data-duration="{(s['end']-s['start'])/30:.12f}" data-track-index="{track(s)}" data-name="{s['id']} · {escape(s['title'])}" width="1920" height="1080" data-layout-allow-overflow aria-label="{escape(s['xem'])}"></canvas>''' for s in scenes)
(P/'index.html').write_text(f'''<!doctype html>
<html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=1920,height=1080"><title>Xem — Make email more human</title>
<script src="assets/vendor/gsap.min.js"></script>
<style>{faces}
*{{box-sizing:border-box;margin:0;padding:0}}html,body{{width:1920px;height:1080px;overflow:hidden;background:#22251f}}#root{{position:relative;width:100%;height:100%;overflow:hidden}}.clip{{position:absolute;inset:0}}canvas{{display:block;width:100%;height:100%}}
</style></head><body>
<main id="root" data-composition-id="xem-brand-film" data-start="0" data-duration="30.033333333333333" data-width="1920" data-height="1080" data-fps="30">
{clips}
<audio id="xem-soundtrack" src="assets/audio/reference-soundtrack.wav" data-start="0" data-duration="30.033333333333333" data-track-index="7" data-volume="1" preload="auto"></audio>
</main>
<script>window.XEM_SCENES={json.dumps(scenes)};</script>
<script src="assets/vendor/carousel-rig.js"></script><script src="assets/vendor/halftone-field.js"></script><script src="assets/film.js"></script>
<script>
window.__timelines=window.__timelines||{{}};
window.buildXemFilm().then(function (film) {{
  const tl=gsap.timeline({{paused:true}});
  tl.to(film.driver,{{t:30.033333333333333,duration:30.033333333333333,ease:"none"}},0);
  window.XEM_SCENES.forEach(function(s){{tl.addLabel(s.id+" · "+s.title,s.start/30);}});
  window.__timelines["xem-brand-film"]=tl;
}}).catch(function(error){{console.error("Xem film initialization failed",error);throw error;}});
</script></body></html>
''')
brief=(P/'BRIEF.md').read_text().replace('phase: planning','phase: production').replace('plan_approved: false','plan_approved: true')
brief=brief.replace('This first deliverable is the complete scene plan. Animation and the final render follow review of that plan.', 'The scene plan was approved by the user. The final video render is authorized.')
brief=brief.replace('Hyperframes 0.8.78 is scaffolded locally. The generated `index.html` is still the blank starter; it is not a film preview. No production website code has been changed, and no video has been rendered or published.', 'Hyperframes 0.8.78 is the pinned renderer. Production uses 28 timed canvas scenes, source-derived registry effect adapters, local Xem fonts/assets, and the supplied reference soundtrack. No production website code is changed.')
(P/'BRIEF.md').write_text(brief)
print('Built 28 labelled Studio scenes, 901-frame timeline, and audio track.')
