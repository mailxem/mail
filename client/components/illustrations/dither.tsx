"use client";
import { useEffect, useRef } from "react";

/**
 * Generated dither illustrations. A density map (0 = empty, 1 = solid) is
 * drawn as a grid of squares whose size follows the density. Everything comes
 * from a fixed seed, so each scene looks the same on every load.
 */
type Noise = (x: number, y: number) => number;
type Field = (nx: number, ny: number, n: Noise, aspect: number) => number;

function noise(seed: number): Noise {
  const hash = (x: number, y: number) => {
    let h = (x * 374761393 + y * 668265263 + seed * 2147483647) | 0;
    h = Math.imul(h ^ (h >>> 13), 1274126177);
    return ((h ^ (h >>> 16)) >>> 0) / 4294967295;
  };
  const smooth = (t: number) => t * t * (3 - 2 * t);
  const value = (x: number, y: number) => {
    const xi = Math.floor(x), yi = Math.floor(y);
    const u = smooth(x - xi), v = smooth(y - yi);
    const a = hash(xi, yi), b = hash(xi + 1, yi), c = hash(xi, yi + 1), d = hash(xi + 1, yi + 1);
    return a + (b - a) * u + (c - a) * v + (a - b - c + d) * u * v;
  };
  return (x, y) => value(x, y) * 0.55 + value(x * 2.1, y * 2.1) * 0.3 + value(x * 4.3, y * 4.3) * 0.15;
}
const clamp = (v: number) => Math.min(1, Math.max(0, v));
const ease = (t: number) => {
  const k = clamp(t);
  return k * k * (3 - 2 * k);
};

/** Dense sky above, white clouds through the upper middle, dotted haze below. */
export const skyField: Field = (nx, ny, n, aspect) => {
  const x = nx * aspect;
  const band = 0.42 - Math.abs(ny - 0.4) * 1.1;
  const cloud = ease((n(x * 4, ny * 6) + band - 0.55) / 0.22);
  const haze = 0.6 + n(x * 10, ny * 12) * 0.35;
  const sky = 0.95 + (haze - 0.95) * ease((ny - 0.3) / 0.3);
  return Math.max(0, sky - cloud);
};

export function Dither({
  field,
  ink,
  inkDark = ink,
  cell = 5,
  seed = 7,
  className,
}: {
  field: Field;
  ink: string;
  inkDark?: string;
  cell?: number;
  seed?: number;
  className?: string;
}) {
  const ref = useRef<HTMLCanvasElement>(null);
  useEffect(() => {
    const canvas = ref.current;
    if (!canvas) return;
    const n = noise(seed);
    const draw = () => {
      const { width, height } = canvas.getBoundingClientRect();
      if (!width || !height) return;
      const dpr = window.devicePixelRatio || 1;
      canvas.width = Math.round(width * dpr);
      canvas.height = Math.round(height * dpr);
      const ctx = canvas.getContext("2d");
      if (!ctx) return;
      ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
      ctx.clearRect(0, 0, width, height);
      ctx.fillStyle = document.documentElement.classList.contains("dark") ? inkDark : ink;
      const aspect = width / height;
      for (let y = 0; y < height; y += cell) {
        for (let x = 0; x < width; x += cell) {
          const size = Math.round(field(x / width, y / height, n, aspect) * (cell - 1));
          if (size > 0) ctx.fillRect(x + (cell - size) / 2, y + (cell - size) / 2, size, size);
        }
      }
    };
    draw();
    const resize = new ResizeObserver(draw);
    resize.observe(canvas);
    // Redraw when the theme class on <html> changes.
    const theme = new MutationObserver(draw);
    theme.observe(document.documentElement, { attributes: true, attributeFilter: ["class"] });
    return () => {
      resize.disconnect();
      theme.disconnect();
    };
  }, [field, ink, inkDark, cell, seed]);
  return <canvas ref={ref} className={className} aria-hidden="true" />;
}
