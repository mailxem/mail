// Shader and uniform mapping from installed HyperFrames Halftone Field.

      window.createXemHalftone = function (canvas) {
        var DUR = 10;
        var W = 1920;
        var H = 1080;

        var settings = {frequency: 2, speed: 1, cellSize: 14, gamma: 10, paletteBias: 1.5,
          palette1: '#5b3cc4', palette2: '#e7d8fa', palette3: '#edf09b', palette4: '#5b3cc4', background: '#22251f'};
        function raw(id) { return settings[id]; }
        function num(id, fallback) { return settings[id] == null ? fallback : settings[id]; }
        function bool(id) { return false; }
        // #RGB / #RRGGBB -> [r, g, b] in 0..1. Anything else falls back.
        function rgb(id, fallback) {
          var s = String(raw(id) || "").trim();
          var m = /^#([0-9a-f]{3}|[0-9a-f]{6})$/i.exec(s);
          if (!m) s = fallback;
          else s = "#" + m[1];
          var hex = s.slice(1);
          if (hex.length === 3) {
            hex = hex[0] + hex[0] + hex[1] + hex[1] + hex[2] + hex[2];
          }
          var n = parseInt(hex, 16);
          return [((n >> 16) & 255) / 255, ((n >> 8) & 255) / 255, (n & 255) / 255];
        }
        // The reference exposes 1-10 / 1-100 / 1-20 dials and maps them onto
        // the ranges the shader actually wants.
        function mapLinear(x, a1, a2, b1, b2) {
          return b1 + ((x - a1) * (b2 - b1)) / (a2 - a1);
        }

        var FREQ = mapLinear(num("frequency", 2), 1, 10, 0.3, 6);
        var SPEED = num("speed", 5) * 0.05;
        var CELL = mapLinear(num("cellSize", 54), 1, 100, 6, 60);
        var GAMMA = mapLinear(num("gamma", 12), 1, 20, 0.5, 8);
        var BIAS = num("paletteBias", 2) * 0.05;
        var PAL = [
          rgb("palette1", "#00AAFF"),
          rgb("palette2", "#C2FF67"),
          rgb("palette3", "#6600FF"),
          rgb("palette4", "#000000"),
        ];
        var BG = rgb("background", "#000000");
        var QUANTIZE = bool("quantize");

        var gl =
          canvas.getContext("webgl", {
            alpha: false,
            antialias: false,
            depth: false,
            stencil: false,
            preserveDrawingBuffer: true,
            powerPreference: "high-performance",
          }) ||
          canvas.getContext("experimental-webgl", {
            alpha: false,
            preserveDrawingBuffer: true,
          });

        var VERT = "attribute vec2 aPos;\nvoid main() { gl_Position = vec4(aPos, 0.0, 1.0); }";

        // Single pass. The reference renders the colour field into a target and
        // samples it at each cell centre, but the field is analytic, so
        // evaluating it at the cell centre here is the same value without the
        // round trip through a texture.
        var FRAG = `
precision highp float;
uniform vec2 uRes;
uniform float uTime;
uniform float uFreq;
uniform float uSpeed;
uniform float uCell;
uniform float uGamma;
uniform float uBias;
uniform vec3 uPal0;
uniform vec3 uPal1;
uniform vec3 uPal2;
uniform vec3 uPal3;
uniform vec3 uBg;

// 3D simplex noise — Ashima Arts / Stefan Gustavson (MIT).
vec3 mod289(vec3 x) { return x - floor(x * (1.0 / 289.0)) * 289.0; }
vec4 mod289(vec4 x) { return x - floor(x * (1.0 / 289.0)) * 289.0; }
vec4 permute(vec4 x) { return mod289(((x * 34.0) + 1.0) * x); }
vec4 taylorInvSqrt(vec4 r) { return 1.79284291400159 - 0.85373472095314 * r; }

float snoise(vec3 v) {
  const vec2 C = vec2(1.0 / 6.0, 1.0 / 3.0);
  const vec4 D = vec4(0.0, 0.5, 1.0, 2.0);

  vec3 i = floor(v + dot(v, C.yyy));
  vec3 x0 = v - i + dot(i, C.xxx);

  vec3 g = step(x0.yzx, x0.xyz);
  vec3 l = 1.0 - g;
  vec3 i1 = min(g.xyz, l.zxy);
  vec3 i2 = max(g.xyz, l.zxy);

  vec3 x1 = x0 - i1 + C.xxx;
  vec3 x2 = x0 - i2 + C.yyy;
  vec3 x3 = x0 - D.yyy;

  i = mod289(i);
  vec4 p = permute(permute(permute(
      i.z + vec4(0.0, i1.z, i2.z, 1.0)) +
      i.y + vec4(0.0, i1.y, i2.y, 1.0)) +
      i.x + vec4(0.0, i1.x, i2.x, 1.0));

  float n_ = 0.142857142857;
  vec3 ns = n_ * D.wyz - D.xzx;

  vec4 j = p - 49.0 * floor(p * ns.z * ns.z);

  vec4 x_ = floor(j * ns.z);
  vec4 y_ = floor(j - 7.0 * x_);

  vec4 x = x_ * ns.x + ns.yyyy;
  vec4 y = y_ * ns.x + ns.yyyy;
  vec4 h = 1.0 - abs(x) - abs(y);

  vec4 b0 = vec4(x.xy, y.xy);
  vec4 b1 = vec4(x.zw, y.zw);

  vec4 s0 = floor(b0) * 2.0 + 1.0;
  vec4 s1 = floor(b1) * 2.0 + 1.0;
  vec4 sh = -step(h, vec4(0.0));

  vec4 a0 = b0.xzyw + s0.xzyw * sh.xxyy;
  vec4 a1 = b1.xzyw + s1.xzyw * sh.zzww;

  vec3 p0 = vec3(a0.xy, h.x);
  vec3 p1 = vec3(a0.zw, h.y);
  vec3 p2 = vec3(a1.xy, h.z);
  vec3 p3 = vec3(a1.zw, h.w);

  vec4 norm = taylorInvSqrt(vec4(dot(p0, p0), dot(p1, p1), dot(p2, p2), dot(p3, p3)));
  p0 *= norm.x;
  p1 *= norm.y;
  p2 *= norm.z;
  p3 *= norm.w;

  vec4 m = max(0.6 - vec4(dot(x0, x0), dot(x1, x1), dot(x2, x2), dot(x3, x3)), 0.0);
  m = m * m;
  return 42.0 * dot(m * m, vec4(dot(p0, x0), dot(p1, x1), dot(p2, x2), dot(p3, x3)));
}

// Full-saturation, full-value HSV: only the hue ramp survives.
vec3 hue2rgb(float h) {
  return clamp(abs(mod(h * 6.0 + vec3(0.0, 4.0, 2.0), 6.0) - 3.0) - 1.0, 0.0, 1.0);
}

// Piecewise-linear ramp across the four palette stops.
vec3 ramp(float v) {
  float x = clamp(v, 0.0, 1.0) * 3.0;
  vec3 c = mix(uPal0, uPal1, clamp(x, 0.0, 1.0));
  c = mix(c, uPal2, clamp(x - 1.0, 0.0, 1.0));
  c = mix(c, uPal3, clamp(x - 2.0, 0.0, 1.0));
  return c;
}

void main() {
  // One sample per cell, taken at the cell centre.
  vec2 cell = floor(gl_FragCoord.xy / uCell);
  vec2 centre = (cell + 0.5) * uCell;

  vec2 st = centre / uRes;
  st.x *= uRes.x / uRes.y;

  float hue = abs(snoise(vec3(st * uFreq, uTime * uSpeed)));
  vec3 field = hue2rgb(hue);

  // The gamma is the whole look: it crushes the midtones so most of the frame
  // stays background and only the crests of the field open up.
  float gray = pow(dot(field, vec3(0.2126, 0.7152, 0.0722)), uGamma);
  float v = clamp(gray + uBias, 0.0, 1.0);

  float radius = v * 0.5 * uCell;
  float d = length(gl_FragCoord.xy - centre);
  float mask = 1.0 - smoothstep(radius - 1.0, radius + 1.0, d);

  gl_FragColor = vec4(mix(uBg, ramp(v), mask), 1.0);
}
`;

        var uni = {};
        var ready = false;

        function compile(type, src) {
          var sh = gl.createShader(type);
          gl.shaderSource(sh, src);
          gl.compileShader(sh);
          if (!gl.getShaderParameter(sh, gl.COMPILE_STATUS)) {
            throw new Error("halftone-field shader: " + gl.getShaderInfoLog(sh));
          }
          return sh;
        }

        if (gl) {
          var prog = gl.createProgram();
          gl.attachShader(prog, compile(gl.VERTEX_SHADER, VERT));
          gl.attachShader(prog, compile(gl.FRAGMENT_SHADER, FRAG));
          gl.linkProgram(prog);
          if (!gl.getProgramParameter(prog, gl.LINK_STATUS)) {
            throw new Error("halftone-field link: " + gl.getProgramInfoLog(prog));
          }
          gl.useProgram(prog);

          var buf = gl.createBuffer();
          gl.bindBuffer(gl.ARRAY_BUFFER, buf);
          gl.bufferData(gl.ARRAY_BUFFER, new Float32Array([-1, -1, 3, -1, -1, 3]), gl.STATIC_DRAW);
          var loc = gl.getAttribLocation(prog, "aPos");
          gl.enableVertexAttribArray(loc);
          gl.vertexAttribPointer(loc, 2, gl.FLOAT, false, 0, 0);

          [
            "uRes",
            "uTime",
            "uFreq",
            "uSpeed",
            "uCell",
            "uGamma",
            "uBias",
            "uPal0",
            "uPal1",
            "uPal2",
            "uPal3",
            "uBg",
          ].forEach(function (n) {
            uni[n] = gl.getUniformLocation(prog, n);
          });

          gl.viewport(0, 0, W, H);
          gl.uniform2f(uni.uRes, W, H);
          gl.uniform1f(uni.uFreq, FREQ);
          gl.uniform1f(uni.uSpeed, SPEED);
          gl.uniform1f(uni.uCell, CELL);
          gl.uniform1f(uni.uGamma, GAMMA);
          gl.uniform1f(uni.uBias, BIAS);
          gl.uniform3f(uni.uPal0, PAL[0][0], PAL[0][1], PAL[0][2]);
          gl.uniform3f(uni.uPal1, PAL[1][0], PAL[1][1], PAL[1][2]);
          gl.uniform3f(uni.uPal2, PAL[2][0], PAL[2][1], PAL[2][2]);
          gl.uniform3f(uni.uPal3, PAL[3][0], PAL[3][1], PAL[3][2]);
          gl.uniform3f(uni.uBg, BG[0], BG[1], BG[2]);
          ready = true;
        }

        // Every frame is computed from t alone — the field is a closed-form
        // function of position and time. No accumulator, no clock, no PRNG.
        // `quantize` snaps t to a 30fps grid for the stepped LED-wall look.
        function draw(t) {
          if (!ready) return;
          gl.uniform1f(uni.uTime, QUANTIZE ? Math.floor(t * 30) / 30 : t);
          gl.drawArrays(gl.TRIANGLES, 0, 3);
          gl.flush();
        }

        draw(0);
        return {draw: draw, canvas: canvas, ready: ready};
      };
