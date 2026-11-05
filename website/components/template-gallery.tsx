"use client";

import { useState } from "react";
import Image from "next/image";
import Link from "next/link";
import { galleryCategories, galleryTemplates } from "@/lib/template-gallery";

export function TemplateGallery() {
  const [category, setCategory] = useState("All");
  const shown = galleryTemplates.filter(
    (t) => category === "All" || t.category === category,
  );
  return (
    <>
      <div
        role="group"
        aria-label="Filter templates by category"
        className="flex flex-wrap gap-2"
      >
        {["All", ...galleryCategories].map((c) => (
          <button
            key={c}
            aria-pressed={category === c}
            onClick={() => setCategory(c)}
            className={`rounded-full border px-4 py-2 text-sm transition-colors ${category === c ? "border-iris bg-iris text-white" : "border-ink/20 hover:border-iris hover:text-iris"}`}
          >
            {c}{" "}
            <span className="opacity-70">
              {c === "All"
                ? galleryTemplates.length
                : galleryTemplates.filter((t) => t.category === c).length}
            </span>
          </button>
        ))}
      </div>
      <ul className="mt-10 grid grid-cols-2 gap-x-5 gap-y-10 md:grid-cols-3 lg:grid-cols-4">
        {shown.map((t) => (
          <li key={t.key}>
            <Link href={`/templates/${t.key}`} className="group block">
              <div className="overflow-hidden rounded-xl border border-ink/10 bg-lavender p-3 pb-0 transition-transform duration-300 group-hover:-translate-y-1 motion-reduce:transform-none">
                <Image
                  src={`/images/gallery/${t.key}.webp`}
                  alt={`Preview of the ${t.name} email template`}
                  width={400}
                  height={500}
                  loading="lazy"
                  className="h-auto w-full rounded-t shadow-md"
                />
              </div>
              <h2 className="mt-4 font-editorial text-xl leading-tight tracking-tight group-hover:text-iris">
                {t.name}
              </h2>
              <p className="mt-1 text-[10px] uppercase tracking-widest text-ink/70">
                {t.category}
              </p>
            </Link>
          </li>
        ))}
      </ul>
    </>
  );
}
