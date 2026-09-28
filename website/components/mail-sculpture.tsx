"use client";
import { useEffect, useRef } from "react";
import * as THREE from "three";

// A single lightweight ribbon: no external textures, no postprocessing, no tracking.
export default function MailSculpture({ paused }: { paused: boolean }) {
  const host = useRef<HTMLDivElement>(null);
  const pause = useRef(paused);
  useEffect(() => {
    pause.current = paused;
  }, [paused]);
  useEffect(() => {
    const container = host.current;
    if (!container) return;
    const reduced = window.matchMedia("(prefers-reduced-motion: reduce)");
    if (reduced.matches) return;
    let renderer: THREE.WebGLRenderer;
    try {
      renderer = new THREE.WebGLRenderer({
        alpha: true,
        antialias: true,
        powerPreference: "low-power",
      });
    } catch {
      return;
    }
    renderer.setPixelRatio(Math.min(window.devicePixelRatio, 1.5));
    renderer.setClearColor(0, 0);
    container.appendChild(renderer.domElement);
    const scene = new THREE.Scene();
    const camera = new THREE.PerspectiveCamera(35, 1, 0.1, 100);
    camera.position.set(0, 0, 15);
    const group = new THREE.Group();
    scene.add(group);
    const points = [];
    const count = 160;
    for (let i = 0; i <= count; i++) {
      const t = i / count;
      const x = (t - 0.5) * 20;
      const y = Math.sin(t * Math.PI * 2 - 0.5) * 1.1;
      const z = Math.sin(t * Math.PI * 3) * 0.7;
      points.push(new THREE.Vector3(x, y, z));
    }
    const curve = new THREE.CatmullRomCurve3(points);
    const positions = [],
      indices = [],
      normals = [];
    for (let i = 0; i <= count; i++) {
      const t = i / count;
      const p = curve.getPointAt(t);
      const tangent = curve.getTangentAt(t);
      const normal = new THREE.Vector3(-tangent.y, tangent.x, 0).normalize();
      const twist = Math.sin(t * Math.PI * 4) * 0.75;
      normal.applyAxisAngle(tangent, twist);
      for (const s of [-1, 1]) {
        const q = p.clone().addScaledVector(normal, s * 0.13);
        positions.push(q.x, q.y, q.z);
        normals.push(0, 0, 1);
      }
      if (i < count) {
        const a = i * 2;
        indices.push(a, a + 1, a + 2, a + 1, a + 3, a + 2);
      }
    }
    const geometry = new THREE.BufferGeometry();
    geometry.setAttribute(
      "position",
      new THREE.Float32BufferAttribute(positions, 3),
    );
    geometry.setIndex(indices);
    geometry.computeVertexNormals();
    const material = new THREE.MeshStandardMaterial({
      color: 0x456749,
      side: THREE.DoubleSide,
      metalness: 0.15,
      roughness: 0.55,
    });
    const ribbon = new THREE.Mesh(geometry, material);
    ribbon.rotation.z = 0.12;
    group.add(ribbon);
    const sphereGeometry = new THREE.IcosahedronGeometry(0.14, 1);
    const sphereMaterial = new THREE.MeshStandardMaterial({
      color: 0xd8aa68,
      metalness: 0.1,
      roughness: 0.4,
    });
    const beads = Array.from({ length: 6 }, (_, i) => {
      const m = new THREE.Mesh(sphereGeometry, sphereMaterial);
      group.add(m);
      return m;
    });
    scene.add(new THREE.AmbientLight(0xffffff, 2));
    const light = new THREE.DirectionalLight(0xfff6d9, 3);
    light.position.set(-3, 4, 8);
    scene.add(light);
    let frame = 0;
    let visible = true;
    let pointerX = 0;
    let pointerY = 0;
    let time = 0;
    let last = 0;
    let disposed = false;
    const resize = () => {
      const { width, height } = container.getBoundingClientRect();
      if (!width || !height) return;
      renderer.setSize(width, height);
      camera.aspect = width / height;
      camera.updateProjectionMatrix();
    };
    const observer = new ResizeObserver(resize);
    observer.observe(container);
    resize();
    const render = (stamp: number) => {
      if (disposed) return;
      frame = requestAnimationFrame(render);
      if (!visible || document.hidden || pause.current) {
        last = stamp;
        return;
      }
      time += Math.min((stamp - last) / 1000, 0.05);
      last = stamp;
      group.rotation.x += (pointerY * 0.1 - group.rotation.x) * 0.025;
      group.rotation.y += (pointerX * 0.09 - group.rotation.y) * 0.025;
      ribbon.position.y = Math.sin(time * 0.35) * 0.08;
      beads.forEach((m, i) => {
        const t = (time * 0.028 + i / 6) % 1;
        m.position.copy(curve.getPointAt(t));
        m.position.applyAxisAngle(new THREE.Vector3(0, 0, 1), 0.12);
        m.rotation.x = time * 0.2;
      });
      renderer.render(scene, camera);
    };
    const intersection = new IntersectionObserver(
      ([e]) => {
        visible = e.isIntersecting;
      },
      { rootMargin: "80px" },
    );
    intersection.observe(container);
    const move = (e: PointerEvent) => {
      pointerX = e.clientX / window.innerWidth - 0.5;
      pointerY = e.clientY / window.innerHeight - 0.5;
    };
    window.addEventListener("pointermove", move, { passive: true });
    const loss = (e: Event) => {
      e.preventDefault();
      container.classList.add("invisible");
      cancelAnimationFrame(frame);
    };
    renderer.domElement.addEventListener("webglcontextlost", loss);
    const motion = () => {
      if (reduced.matches) {
        cancelAnimationFrame(frame);
        renderer.clear();
      } else {
        last = performance.now();
        frame = requestAnimationFrame(render);
      }
    };
    reduced.addEventListener("change", motion);
    renderer.render(scene, camera);
    frame = requestAnimationFrame(render);
    return () => {
      disposed = true;
      cancelAnimationFrame(frame);
      observer.disconnect();
      intersection.disconnect();
      window.removeEventListener("pointermove", move);
      reduced.removeEventListener("change", motion);
      renderer.domElement.removeEventListener("webglcontextlost", loss);
      geometry.dispose();
      material.dispose();
      sphereGeometry.dispose();
      sphereMaterial.dispose();
      renderer.dispose();
      renderer.domElement.remove();
    };
  }, []);
  return (
    <div
      ref={host}
      aria-hidden="true"
      className="pointer-events-none absolute -left-[15%] top-[32%] h-[62%] w-[130%] opacity-70 [&_canvas]:h-full [&_canvas]:w-full"
    />
  );
}
