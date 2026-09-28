"use client";

import { useEffect, useRef, useState } from "react";
import Link from "next/link";
import {
  ArrowUpRight,
  Check,
  Download,
  Eye,
  Monitor,
  Search,
  Smartphone,
  Sparkles,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Empty } from "@/components/marketing/shared";
import {
  filterTemplateStarters,
  starterCategories,
  starterCollections,
  templateStarters,
  type TemplateStarter,
} from "@/lib/template-starters/registry";

export function StarterGallery() {
  const [category, setCategory] = useState<string>("All templates");
  const [collection, setCollection] = useState("All collections");
  const [visibleCount, setVisibleCount] = useState(12);
  const sentinel = useRef<HTMLDivElement>(null);
  const pageSize = 12;
  const [search, setSearch] = useState("");
  const [preview, setPreview] = useState<TemplateStarter | null>(null);
  const [mobile, setMobile] = useState(false);
  const starters = filterTemplateStarters(category, search, collection);
  const visibleStarters = starters.slice(0, visibleCount);
  const hasMore = visibleCount < starters.length;
  useEffect(() => {
    if (!hasMore || !sentinel.current) return;
    const observer = new IntersectionObserver(([entry]) => {
      if (entry.isIntersecting) setVisibleCount((count) => count + pageSize);
    }, { rootMargin: "500px" });
    observer.observe(sentinel.current);
    return () => observer.disconnect();
  }, [hasMore, visibleCount, category, search, collection]);
  const editorHref = (starter: TemplateStarter) =>
    `/templates/new?starter=${encodeURIComponent(starter.key)}`;

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-center justify-between gap-5 rounded-2xl border border-violet-100 bg-violet-50/60 p-6">
        <div className="max-w-xl">
          <div className="mb-2 flex items-center gap-2 text-xs font-medium text-violet-700">
            <Sparkles className="size-4" />
            THE XEM COLLECTION
          </div>
          <h3 className="text-2xl font-semibold tracking-tight">
            A head start for your next send.
          </h3>
          <p className="mt-2 text-sm leading-6 text-muted-foreground">
            Thoughtful layouts, ready for your voice. Pick a starting point and
            make every detail yours in the editor.
          </p>
        </div>
        <div className="flex flex-col gap-2 text-xs text-muted-foreground">
          {[
            `${templateStarters.length} templates`,
            "Every block is editable",
            "Built with Xem footer included",
          ].map((label) => (
            <span className="flex items-center gap-2" key={label}>
              <Check className="size-3.5 text-violet-600" />
              {label}
            </span>
          ))}
        </div>
      </div>
      <div className="flex flex-wrap items-center justify-between gap-4">
        <div
          role="group"
          aria-label="Template categories"
          className="flex flex-wrap gap-1.5"
        >
          {starterCategories.map((item) => (
            <Button
              key={item}
              size="sm"
              variant={category === item ? "secondary" : "ghost"}
              aria-pressed={category === item}
              onClick={() => { setCategory(item); setVisibleCount(pageSize); }}
              className={
                category === item
                  ? "bg-violet-100 text-violet-800 hover:bg-violet-100"
                  : "text-muted-foreground"
              }
            >
              {item}
            </Button>
          ))}
        </div>
        <div className="relative w-full sm:w-64">
          <Search className="absolute left-3 top-3 size-4 text-muted-foreground" />
          <Input
            aria-label="Search starter templates"
            placeholder="Find your starting point…"
            className="pl-9"
            value={search}
            onChange={(e) => { setSearch(e.target.value); setVisibleCount(pageSize); }}
          />
        </div>
      </div>
      <div className="flex flex-wrap items-center gap-3">
        <label htmlFor="starter-collection" className="text-xs text-muted-foreground">Collection</label>
        <select id="starter-collection" value={collection}
          onChange={(e) => { setCollection(e.target.value); setVisibleCount(pageSize); }}
          className="h-9 rounded-md border border-input bg-white px-3 text-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring">
          {starterCollections.map((item) => <option key={item}>{item}</option>)}
        </select>
      <p className="text-xs text-muted-foreground" aria-live="polite">
        {starters.length} {starters.length === 1 ? "template" : "templates"}
        {category !== "All templates"
          ? ` · ${category}`
          : " for your next idea"}
      </p>
      </div>
      {starters.length === 0 ? (
        <Empty
          title="No templates found"
          description="Try another category or search for a different style."
          action={
            <Button
              variant="outline"
              onClick={() => {
                setSearch("");
                setCategory("All templates");
                setCollection("All collections");
                setVisibleCount(pageSize);
              }}
            >
              Clear filters
            </Button>
          }
        />
      ) : (
        <div className="grid grid-cols-1 gap-5 md:grid-cols-2 xl:grid-cols-3">
          {visibleStarters.map((starter) => (
            <article
              key={starter.key}
              className="group overflow-hidden rounded-2xl border border-border bg-white transition-shadow hover:shadow-md"
            >
              <button
                type="button"
                aria-label={`Preview ${starter.name}`}
                className="relative block h-[300px] w-full overflow-hidden border-b border-border bg-muted/40 text-left focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-violet-500"
                onClick={() => {
                  setPreview(starter);
                  setMobile(false);
                }}
              >
                <iframe
                  src={starter.previewUrl}
                  title={`${starter.name} thumbnail`}
                  sandbox=""
                  loading="lazy"
                  tabIndex={-1}
                  aria-hidden="true"
                  referrerPolicy="no-referrer"
                  className="pointer-events-none absolute left-1/2 top-0 h-[1100px] w-[600px] origin-top -translate-x-1/2 scale-50 border-0"
                />
                <span className="absolute inset-x-0 bottom-0 flex justify-center bg-gradient-to-t from-black/20 to-transparent p-5 opacity-0 transition-opacity group-hover:opacity-100 group-focus-within:opacity-100">
                  <span className="flex items-center gap-2 rounded-full bg-white px-4 py-2 text-xs font-medium shadow-sm">
                    <Eye className="size-3.5" />
                    Preview design
                  </span>
                </span>
              </button>
              <div className="flex flex-col gap-3 p-5">
                <span className="text-[11px] font-medium uppercase tracking-wider text-violet-600">
                  {starter.category}
                </span>
                <div>
                  <h3 className="text-base font-semibold">{starter.name}</h3>
                  <p className="mt-1.5 min-h-10 text-sm leading-5 text-muted-foreground">
                    {starter.description}
                  </p>
                </div>
                <div className="mt-1 flex gap-2">
                  <Button
                    variant="outline"
                    onClick={() => {
                      setPreview(starter);
                      setMobile(false);
                    }}
                  >
                    Preview
                  </Button>
                  <Button asChild className="flex-1">
                    <Link href={editorHref(starter)}>
                      Use template
                      <ArrowUpRight className="size-4" />
                    </Link>
                  </Button>
                </div>
              </div>
            </article>
          ))}
        </div>
      )}
      <div ref={sentinel} className="flex min-h-12 items-center justify-center" aria-live="polite">
        {hasMore ? <Button variant="ghost" onClick={() => setVisibleCount((count) => count + pageSize)}>Load more templates</Button>
          : starters.length > 0 ? <p className="text-xs text-muted-foreground">You’ve seen all {starters.length} matching templates.</p> : null}
      </div>
      <Dialog
        open={!!preview}
        onOpenChange={(open) => {
          if (!open) setPreview(null);
        }}
      >
        <DialogContent className="flex h-[92dvh] max-w-4xl flex-col gap-0 overflow-hidden p-0">
          {preview && (
            <>
              <DialogHeader className="shrink-0 border-b border-border px-6 pb-4 pt-6 pr-12">
                <DialogTitle>{preview.name}</DialogTitle>
                <DialogDescription>{preview.description}</DialogDescription>
              </DialogHeader>
              <div className="flex shrink-0 flex-wrap items-center justify-between gap-3 border-b border-border bg-white px-6 py-3">
                <div className="min-w-0">
                  <span className="text-[10px] uppercase tracking-wide text-muted-foreground">
                    Subject
                  </span>
                  <p className="text-sm font-medium">{preview.subject}</p>
                </div>
                <div
                  className="flex gap-1"
                  role="group"
                  aria-label="Preview device"
                >
                  <Button
                    variant={mobile ? "ghost" : "secondary"}
                    size="icon"
                    aria-label="Desktop preview"
                    aria-pressed={!mobile}
                    onClick={() => setMobile(false)}
                  >
                    <Monitor className="size-4" />
                  </Button>
                  <Button
                    variant={mobile ? "secondary" : "ghost"}
                    size="icon"
                    aria-label="Mobile preview"
                    aria-pressed={mobile}
                    onClick={() => setMobile(true)}
                  >
                    <Smartphone className="size-4" />
                  </Button>
                </div>
              </div>
              <div className="flex min-h-0 flex-1 justify-center bg-muted/50 p-3 sm:p-5">
                <iframe
                  src={preview.previewUrl}
                  title={`${preview.name} email preview`}
                  sandbox=""
                  referrerPolicy="no-referrer"
                  className={`h-full w-full border-0 bg-white shadow-sm ${mobile ? "max-w-[375px]" : "max-w-[600px]"}`}
                />
              </div>
              <div className="flex shrink-0 flex-wrap items-center justify-between gap-3 border-t border-border bg-white px-6 py-4">
                <Button variant="ghost" asChild>
                  <a href={preview.designUrl} download={`${preview.key}.json`}>
                    <Download className="size-4" />
                    Design JSON
                  </a>
                </Button>
                <Button asChild>
                  <Link href={editorHref(preview)}>
                    Customize in editor
                    <ArrowUpRight className="size-4" />
                  </Link>
                </Button>
              </div>
            </>
          )}
        </DialogContent>
      </Dialog>
    </div>
  );
}
