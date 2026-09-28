// Adapted from installed HyperFrames Circle Carousel 4.

      (function (global) {
        "use strict";

        const lerp = (a, b, t) => a + (b - a) * t;
        const clamp01 = (t) => (t < 0 ? 0 : t > 1 ? 1 : t);

        const aeLinear = (v, a, b, c, d) => (a === b ? c : lerp(c, d, clamp01((v - a) / (b - a))));

        const aeEase = (v, a, b, c, d) => {
          const t = clamp01((v - a) / (b - a));
          const e = t < 0.5 ? 4 * t * t * t : 1 - Math.pow(-2 * t + 2, 3) / 2;
          return lerp(c, d, e);
        };
        const wrap01 = (x) => {
          let p = x % 1;
          if (p < 0) p += 1;
          return p;
        };
        const DEG = 180 / Math.PI;

        function buildPath(vertices, closed, samplesPerSeg, center) {
          samplesPerSeg = samplesPerSeg || 64;
          const n = vertices.length;
          const segCount = closed ? n : n - 1;
          const pts = [];
          for (let s = 0; s < segCount; s++) {
            const a = vertices[s],
              b = vertices[(s + 1) % n];
            const p0 = a.p,
              p1 = [a.p[0] + a.out[0], a.p[1] + a.out[1]];
            const p2 = [b.p[0] + b.in[0], b.p[1] + b.in[1]],
              p3 = b.p;
            const last = s === segCount - 1 ? samplesPerSeg : samplesPerSeg - 1;
            for (let k = 0; k <= last; k++) {
              const t = k / samplesPerSeg,
                mt = 1 - t;
              const x =
                mt * mt * mt * p0[0] +
                3 * mt * mt * t * p1[0] +
                3 * mt * t * t * p2[0] +
                t * t * t * p3[0];
              const y =
                mt * mt * mt * p0[1] +
                3 * mt * mt * t * p1[1] +
                3 * mt * t * t * p2[1] +
                t * t * t * p3[1];
              pts.push([x, y]);
            }
          }
          if (center) {
            let x0 = Infinity,
              x1 = -Infinity,
              y0 = Infinity,
              y1 = -Infinity;
            for (const q of pts) {
              x0 = Math.min(x0, q[0]);
              x1 = Math.max(x1, q[0]);
              y0 = Math.min(y0, q[1]);
              y1 = Math.max(y1, q[1]);
            }
            const sx = center[0] - (x0 + x1) / 2,
              sy = center[1] - (y0 + y1) / 2;
            for (const q of pts) {
              q[0] += sx;
              q[1] += sy;
            }
          }
          const cum = [0];
          for (let i = 1; i < pts.length; i++) {
            const dx = pts[i][0] - pts[i - 1][0],
              dy = pts[i][1] - pts[i - 1][1];
            cum.push(cum[i - 1] + Math.hypot(dx, dy));
          }
          const total = cum[cum.length - 1];
          function locate(pct) {
            const target = pct * total;
            let lo = 0,
              hi = cum.length - 1;
            while (hi - lo > 1) {
              const mid = (lo + hi) >> 1;
              if (cum[mid] <= target) lo = mid;
              else hi = mid;
            }
            const segLen = cum[hi] - cum[lo] || 1;
            return { i: lo, j: hi, f: (target - cum[lo]) / segLen };
          }
          return {
            closed,
            length: total,
            pointOnPath(pct) {
              pct = closed ? wrap01(pct) : clamp01(pct);
              const { i, j, f } = locate(pct);
              return [lerp(pts[i][0], pts[j][0], f), lerp(pts[i][1], pts[j][1], f)];
            },
            tangentOnPath(pct) {
              pct = closed ? wrap01(pct) : clamp01(pct);
              const { i, j } = locate(pct);
              const dx = pts[j][0] - pts[i][0],
                dy = pts[j][1] - pts[i][1];
              const l = Math.hypot(dx, dy) || 1;
              return [dx / l, dy / l];
            },
          };
        }

        function pathRig(path, ctl, comp, n, u) {
          const spc = ctl.spacing / 100;
          const phase = ctl.phaseOffset / 100;
          const manual = ctl.moveCarousel / 100;
          const dAng = (ctl.depthAngle || 0) / DEG;
          const rotAmt = (ctl.rotationFollow || 0) / 100;
          const cx = comp.width / 2,
            cy = comp.height / 2;
          const maxDist = Math.hypot(cx, cy);
          const out = new Array(n);
          for (let idx = 0; idx < n; idx++) {
            const off = (idx / n) * spc + phase;
            const pct = wrap01(u + off + manual);
            const pos = path.pointOnPath(pct);

            const dx = pos[0] - cx,
              dy = pos[1] - cy;
            const projected = dx * Math.sin(dAng) + dy * Math.cos(dAng);
            const norm = (projected / maxDist + 1) / 2;
            const depthScale = aeLinear(norm, 0, 1, ctl.scaleMin, 100);
            const mixed = aeLinear(ctl.scaleDepth / 100, 0, 1, 100, depthScale);
            const scale = (mixed * (ctl.globalScale / 100)) / 100;

            const tang = path.tangentOnPath(pct);
            const rotZ = Math.atan2(tang[1], tang[0]) * DEG * rotAmt;

            let x = pos[0],
              y = pos[1],
              z = 0;
            if (ctl.parent) {
              const rx = (ctl.xSlider || 0) / DEG,
                ry = (ctl.ySlider || 0) / DEG;
              const a = ctl.parent.anchor,
                P = ctl.parent.position;

              let l = [pos[0] - a[0], pos[1] - a[1], 0];
              l = [
                l[0] * Math.cos(ry) + l[2] * Math.sin(ry),
                l[1],
                -l[0] * Math.sin(ry) + l[2] * Math.cos(ry),
              ];
              l = [
                l[0],
                l[1] * Math.cos(rx) - l[2] * Math.sin(rx),
                l[1] * Math.sin(rx) + l[2] * Math.cos(rx),
              ];
              x = P[0] + l[0];
              y = P[1] + l[1];
              z = (P[2] || 0) + l[2];
            }

            let rotY = 0,
              rotX = 0;
            if (ctl.parent) {
              const camZ = -(comp.cameraDistance || comp.cameraZoom || 5669);
              const vx = cx - x,
                vy = cy - y,
                vz = camZ - z;
              rotY = Math.atan2(vx, -vz) * DEG;
              rotX = -Math.atan2(vy, Math.hypot(vx, vz)) * DEG;
            }
            if (ctl.yAngle3d) {
              const dist = Math.abs(x - cx);
              let r = aeEase(dist, 0, comp.width * (700 / 1920), 0, ctl.yAngle3d);
              if (ctl.mirrorY !== false && x < cx) r = -r;
              rotY += r;
            }
            out[idx] = { x, y, z, scale, rotZ, rotY, rotX, pct, depth: norm };
          }
          return out;
        }

        function orbitRig(ctl, comp, n, u) {
          const r = ctl.radius;
          const rawCnt = Math.max(1, ctl.elementCount);
          const cnt = Math.floor(rawCnt);
          const off = ctl.offset;

          const st = ctl.spinTurns || { x: 0, y: 1, z: 0 };
          const rx = ((ctl.rotationX || 0) + u * 360 * (st.x || 0)) / DEG;
          const ry = ((ctl.rotationY || 0) + u * 360 * (st.y || 0)) / DEG;
          const rz = ((ctl.rotationZ || 0) + u * 360 * (st.z || 0)) / DEG;
          const cx = comp.width / 2,
            cy = comp.height / 2;
          const camZ = -(comp.cameraDistance || comp.cameraZoom || 2000);
          const cd = Math.abs(camZ);
          const ga = Math.PI * (3 - Math.sqrt(5));
          const getP = (index, c) => {
            const i = (((index - 0.5 + off) % c) + c) % c;
            const inc = Math.acos(Math.max(-1, Math.min(1, 1 - 2 * (i / c))));
            const azi = ga * i;
            return [
              Math.sin(inc) * Math.cos(azi) * r,
              Math.cos(inc) * r,
              Math.sin(inc) * Math.sin(azi) * r,
            ];
          };
          const out = new Array(n);
          for (let k = 0; k < n; k++) {
            const index = k + 1;
            const a = getP(index, cnt),
              b = getP(index, cnt + 1),
              f = rawCnt % 1;
            let p = [lerp(a[0], b[0], f), lerp(a[1], b[1], f), lerp(a[2], b[2], f)];

            p = [
              p[0],
              p[1] * Math.cos(rx) - p[2] * Math.sin(rx),
              p[1] * Math.sin(rx) + p[2] * Math.cos(rx),
            ];
            p = [
              p[0] * Math.cos(ry) + p[2] * Math.sin(ry),
              p[1],
              -p[0] * Math.sin(ry) + p[2] * Math.cos(ry),
            ];
            p = [
              p[0] * Math.cos(rz) - p[1] * Math.sin(rz),
              p[0] * Math.sin(rz) + p[1] * Math.cos(rz),
              p[2],
            ];
            const jitter = 1 + (index % 2) * 0.01;
            const x = cx + p[0] * jitter,
              y = cy + p[1] * jitter,
              z = p[2] * jitter;

            const d = Math.hypot(x - cx, y - cy, z - camZ);
            const t = clamp01(aeLinear(d, cd - r, cd + r, 0, 1));
            const s = aeLinear(t, 0, 1, ctl.scaleMax, ctl.scaleMin);
            const finalS = aeLinear(ctl.perspectiveBoost, 0, 100, 100, s);

            const behindCam = z <= camZ * 0.92;

            const vx = cx - x,
              vy = cy - y,
              vz = camZ - z;
            const yaw = Math.atan2(vx, -vz) * DEG;
            const pitch = -Math.atan2(vy, Math.hypot(vx, vz)) * DEG;
            out[k] = {
              x,
              y,
              z,
              scale: finalS / 100,
              rotZ: 0,
              rotY: yaw,
              rotX: pitch,
              visible: index <= cnt && !behindCam,
              depth: 1 - t,
            };
          }
          return out;
        }

        function aeCurve(keys) {
          const segs = [];
          for (let i = 0; i < keys.length - 1; i++) {
            const a = keys[i],
              b = keys[i + 1],
              T = b.time - a.time;
            const linear = !a.easeOut || !b.easeIn || (a.flags && a.flags.startsWith("0101"));
            let P1, P2;
            if (linear) {
              P1 = [a.time + T / 3, a.value + (b.value - a.value) / 3];
              P2 = [a.time + (2 * T) / 3, a.value + (2 * (b.value - a.value)) / 3];
            } else {
              const lo = a.easeOut.influence * T,
                li = b.easeIn.influence * T;
              P1 = [a.time + lo, a.value + a.easeOut.speed * lo];
              P2 = [b.time - li, b.value - b.easeIn.speed * li];
            }
            segs.push({
              t0: a.time,
              t1: b.time,
              P0: [a.time, a.value],
              P1,
              P2,
              P3: [b.time, b.value],
            });
          }
          const bez = (p0, p1, p2, p3, s) => {
            const m = 1 - s;
            return m * m * m * p0 + 3 * m * m * s * p1 + 3 * m * s * s * p2 + s * s * s * p3;
          };
          return {
            first: keys[0].value,
            last: keys[keys.length - 1].value,
            duration: keys[keys.length - 1].time - keys[0].time,
            at(t) {
              if (t <= segs[0].t0) return keys[0].value;
              const sg = segs.find((g) => t <= g.t1) || segs[segs.length - 1];
              if (t >= sg.t1) return sg.P3[1];

              let s = (t - sg.t0) / (sg.t1 - sg.t0);
              for (let k = 0; k < 8; k++) {
                const x = bez(sg.P0[0], sg.P1[0], sg.P2[0], sg.P3[0], s) - t;
                const m = 1 - s;
                const dx =
                  3 * m * m * (sg.P1[0] - sg.P0[0]) +
                  6 * m * s * (sg.P2[0] - sg.P1[0]) +
                  3 * s * s * (sg.P3[0] - sg.P2[0]);
                if (Math.abs(dx) < 1e-9) break;
                s -= x / dx;
                s = s < 0 ? 0 : s > 1 ? 1 : s;
              }
              return bez(sg.P0[1], sg.P1[1], sg.P2[1], sg.P3[1], s);
            },
          };
        }

        function chainedDrive(keys, mode) {
          const c = aeCurve(keys),
            T = c.duration,
            delta = c.last - c.first;
          return {
            cycle: T,
            at(t) {
              if (T <= 0) return c.first;
              const k = Math.floor(t / T),
                r = t - k * T;
              if (mode === "pingpong")
                return k % 2 === 0 ? c.at(keys[0].time + r) : c.at(keys[0].time + (T - r));
              return c.at(keys[0].time + r) + k * delta;
            },
          };
        }

        function buildProgress(opts) {
          const dur = opts.duration,
            cruise = opts.cruise;
          const sign = cruise < 0 ? -1 : 1;
          const c = Math.abs(cruise);
          const cap = opts.boostCap == null ? 0.5 : opts.boostCap;
          const vIn = Math.max(
            Math.min(c * (opts.arriveMul || 4), c + cap),
            opts.arriveMin || 0.16,
          );
          const vOut = Math.max(Math.min(c * (opts.exitMul || 6), c + cap), opts.exitMin || 0.22);
          const tA = opts.arrive || 1.6,
            tX = opts.exit || 0.9;
          const N = Math.max(600, Math.round(dur * 120));
          const lut = new Float64Array(N + 1);
          let acc = 0,
            prevV = null;
          for (let i = 0; i <= N; i++) {
            const t = (i / N) * dur;
            let v;
            if (t < tA) {
              const k = 1 - Math.pow(1 - t / tA, 4);
              v = lerp(vIn, c, k);
            } else if (t > dur - tX) {
              const k = Math.pow((t - (dur - tX)) / tX, 3);
              v = lerp(c, vOut, k);
            } else v = c;
            if (prevV !== null) acc += ((prevV + v) / 2) * (dur / N);
            prevV = v;
            lut[i] = acc * sign;
          }
          return {
            duration: dur,
            travel: lut[N],
            at(t) {
              const f = clamp01(t / dur) * N,
                i = Math.floor(f),
                g = f - i;
              return i >= N ? lut[N] : lerp(lut[i], lut[i + 1], g);
            },
          };
        }

        global.CarouselRig = {
          buildPath,
          pathRig,
          orbitRig,
          buildProgress,
          aeCurve,
          chainedDrive,
          aeLinear,
          aeEase,
        };
      })(typeof window !== "undefined" ? window : globalThis);


      window.XEM_CAROUSEL_DATA = {
        id: "carousel-circle-4",
        family: "circle",
        controls: {
          speed: 30,
          spacing: 100,
          phaseOffset: 0,
          globalScale: 70.2,
          scaleDepth: 135,
          scaleMin: -27.5,
          depthAngle: 584,
          moveCarousel: 0,
          rotationFollow: 0,
        },
        comp: {
          width: 3840,
          height: 2160,
          outWidth: 1920,
          outHeight: 1080,
          cameraZoom: 5669.29,
          cameraDistance: 5669.29,
          duration: 6,
          fps: 30,
        },
        n: 12,
        cards: [
          { file: "artist-bob-seger.jpg", width: 640, height: 640 },
          { file: "artist-dire-straits.jpg", width: 640, height: 640 },
          { file: "artist-faces.jpg", width: 640, height: 500 },
          { file: "artist-little-feat.jpg", width: 640, height: 640 },
          { file: "artist-north-mississippi-allstars.jpg", width: 640, height: 640 },
          { file: "artist-sea-level.jpg", width: 640, height: 495 },
          { file: "artist-taj-mahal.jpg", width: 640, height: 640 },
          { file: "artist-tool.jpg", width: 640, height: 640 },
          { file: "artist-van-morrison.jpg", width: 640, height: 640 },
          { file: "pl-10-songs.jpg", width: 640, height: 640 },
          { file: "pl-bangers.jpg", width: 640, height: 640 },
          { file: "pl-bhindi-bhagee.jpg", width: 640, height: 640 },
        ],
        path: {
          vertices: [
            { p: [2261.362, 570.612], in: [-520.1, 120.008], out: [520.1, -120.008] },
            { p: [3319.621, 929.734], in: [-64.355, -318.344], out: [64.355, 318.344] },
            { p: [2494.415, 1723.447], in: [520.1, -120.008], out: [-520.1, 120.008] },
            { p: [1436.156, 1364.325], in: [64.355, 318.344], out: [-64.355, -318.344] },
          ],
          closed: true,
        },
        keys: {
          "Move Carousel": [
            {
              time: 0,
              value: 0,
              flags: "02020000",
              easeIn: { speed: 0, influence: 0.166667 },
              easeOut: { speed: 238.51658, influence: 0.124734 },
            },
            {
              time: 1.51667,
              value: 75.5,
              flags: "02020000",
              easeIn: { speed: 563.152726, influence: 0.063927 },
              easeOut: { speed: 0, influence: 0.166667 },
            },
          ],
        },
        motion: {
          arrive: 1.6,
          arriveMul: 4,
          arriveMin: 0.16,
          exit: 0.9,
          exitMul: 6,
          exitMin: 0.22,
          entryStaggerTotal: 0.45,
          entryDuration: 0.55,
          entryLead: 0.25,
          orbitSpinDegPerSec: 24,
          orbitSpinScale: 1,
          boostCap: 0.5,
          speedScale: 1,
        },
        look: {
          background: "#EDEDEF",
          cardRadius: 0,
          cardShadow: "none",
          imageHeight: null,
          centerPath: true,
          endFade: 0.06,
          fitBox: 1500,
        },
        text: null,
        ending: null,
        outro: null,
        ring: null,
        registry: {
          imageVars: [
            "image1",
            "image2",
            "image3",
            "image4",
            "image5",
            "image6",
            "image7",
            "image8",
            "image9",
            "image10",
            "image11",
            "image12",
          ],
          textVars: null,
          speedVar: "speed",
        },
        cutAt: null,
      };
