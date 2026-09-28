"use client";
import { useMemo, useState } from "react";
import { Search, X } from "lucide-react";
import { PostCard } from "./post-card";
import type { PostMeta } from "@/lib/blog-types";
export function BlogExplorer({ posts }: { posts: PostMeta[] }) {
  const [category, setCategory] = useState("All stories");
  const [query, setQuery] = useState("");
  const categories = ["All stories", ...new Set(posts.map((p) => p.category))];
  const filtered = useMemo(
    () =>
      posts.filter(
        (p) =>
          (category === "All stories" || p.category === category) &&
          `${p.title} ${p.description} ${p.tags.join(" ")}`
            .toLowerCase()
            .includes(query.trim().toLowerCase()),
      ),
    [posts, category, query],
  );
  return (
    <section
      aria-labelledby="all-stories"
      className="mx-auto max-w-[1160px] px-5 pb-24 md:px-0"
    >
      <div className="flex flex-wrap items-center justify-between gap-5 border-b border-ink/15 pb-6">
        <h2 id="all-stories" className="font-editorial text-4xl">
          Find your next good idea.
        </h2>
        <label className="flex w-full items-center gap-3 rounded-lg border border-ink/20 bg-white/50 px-4 sm:w-72">
          <Search size={16} className="shrink-0 text-muted" />
          <input
            type="search"
            aria-label="Search articles"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Search the journal"
            className="h-11 min-w-0 flex-1 bg-transparent text-sm outline-none"
          />
        </label>
      </div>
      <div
        aria-label="Article categories"
        className="my-6 flex flex-wrap gap-2"
      >
        {categories.map((c) => (
          <button
            key={c}
            onClick={() => setCategory(c)}
            aria-pressed={category === c}
            className={`min-h-10 rounded-full border px-4 text-xs transition-colors ${category === c ? "border-iris bg-iris text-white" : "border-ink/15 hover:border-iris hover:text-iris"}`}
          >
            {c}
          </button>
        ))}
      </div>
      <p aria-live="polite" className="mb-7 text-xs text-muted">
        {filtered.length} {filtered.length === 1 ? "story" : "stories"}
        {query && ` matching “${query}”`}
      </p>
      {filtered.length ? (
        <div className="grid gap-x-7 gap-y-12 sm:grid-cols-2 lg:grid-cols-3">
          {filtered.map((p) => (
            <PostCard key={p.slug} post={p} />
          ))}
        </div>
      ) : (
        <div className="rounded-2xl border border-dashed border-ink/25 py-16 text-center">
          <h3 className="font-editorial text-3xl">No stories here just yet.</h3>
          <p className="mt-3 text-sm text-muted">
            Try a broader search or another category.
          </p>
          <button
            onClick={() => {
              setQuery("");
              setCategory("All stories");
            }}
            className="mx-auto mt-6 flex items-center gap-2 text-sm text-iris"
          >
            <X size={15} />
            Clear filters
          </button>
        </div>
      )}
    </section>
  );
}
