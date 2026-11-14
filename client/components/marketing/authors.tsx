"use client";
import { useEffect, useRef, useState } from "react";
import { ImagePlus, Pencil, Plus, Trash2, UserRound } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { useApi } from "@/hooks/use-api";
import { useMarketing, useMarketingQuery } from "@/lib/marketing/api";
import type { Author } from "@/lib/marketing/types";
import { workspaceClassName } from "@/lib/workspace-styles";
import { Empty, Field, Modal, QueryState } from "./shared";

const blank: Omit<Author, "id"> = {
  name: "",
  photoUrl: "",
  email: "",
  bylineEmail: true,
  bylineWeb: true,
  isDefault: false,
};

/** The workspace's authors, in Settings. */
export function AuthorsSettings() {
  const q = useMarketingQuery<Author[]>("marketing/authors");
  const { request } = useMarketing();
  const [editing, setEditing] = useState<Partial<Author> | null>(null);
  const remove = async (a: Author) => {
    if (!confirm(`Remove ${a.name}? Campaigns they signed switch to your default author.`)) return;
    try {
      await request(`marketing/authors/${a.id}`, "DELETE");
      await q.refetch();
      toast.success("Author removed");
    } catch (e) {
      toast.error((e as Error).message);
    }
  };
  return (
    <section className={workspaceClassName("product-panel")}>
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h2 className="text-lg font-semibold">Authors</h2>
          <p className="mt-1 text-sm text-muted-foreground">
            Sign campaigns and newsletter issues. Readers see “By” and the
            author’s name at the top of the email and on your newsletter
            website, and replies go to the author.
          </p>
        </div>
        <Button onClick={() => setEditing({ ...blank, isDefault: !q.data?.length })}>
          <Plus className="mr-2 size-4" />
          Add author
        </Button>
      </div>
      <div className="mt-6">
        <QueryState loading={q.isPending} error={q.error} retry={() => void q.refetch()} />
        {q.data?.length === 0 && (
          <Empty
            title="No authors yet"
            description="Add yourself or your writers. The default author signs anything that doesn’t pick one."
          />
        )}
        <ul className="divide-y">
          {q.data?.map((a) => (
            <li key={a.id} className="flex items-center gap-3 py-3">
              <Avatar author={a} size={40} />
              <div className="min-w-0 flex-1">
                <p className="truncate text-sm font-medium">
                  {a.name}
                  {a.isDefault && (
                    <span className="ml-2 rounded-full bg-muted px-2 py-0.5 text-xs font-normal text-muted-foreground">
                      Default
                    </span>
                  )}
                </p>
                <p className="truncate text-xs text-muted-foreground">
                  {[
                    a.email ? `Replies to ${a.email}` : "Replies go to the sender",
                    bylineWhere(a),
                  ].join(" · ")}
                </p>
              </div>
              <Button size="sm" variant="ghost" aria-label={`Edit ${a.name}`} onClick={() => setEditing(a)}>
                <Pencil className="size-4" />
              </Button>
              <Button size="sm" variant="ghost" aria-label={`Remove ${a.name}`} onClick={() => void remove(a)}>
                <Trash2 className="size-4" />
              </Button>
            </li>
          ))}
        </ul>
      </div>
      {editing && (
        <AuthorEditor
          value={editing}
          close={() => setEditing(null)}
          saved={() => void q.refetch()}
        />
      )}
    </section>
  );
}

function bylineWhere(a: Author) {
  if (a.bylineEmail && a.bylineWeb) return "Byline in emails and on the web";
  if (a.bylineEmail) return "Byline in emails only";
  if (a.bylineWeb) return "Byline on the web only";
  return "Byline hidden";
}

export function Avatar({ author, size }: { author: Pick<Author, "name" | "photoUrl">; size: number }) {
  return author.photoUrl ? (
    // eslint-disable-next-line @next/next/no-img-element
    <img src={author.photoUrl} alt="" width={size} height={size} className="shrink-0 rounded-full object-cover" style={{ width: size, height: size }} />
  ) : (
    <span className="flex shrink-0 items-center justify-center rounded-full bg-muted text-muted-foreground" style={{ width: size, height: size }}>
      <UserRound size={size / 2} />
    </span>
  );
}

function AuthorEditor({
  value,
  close,
  saved,
}: {
  value: Partial<Author>;
  close: () => void;
  saved: () => void;
}) {
  const { request } = useMarketing();
  const { apiFetch } = useApi();
  const [draft, setDraft] = useState({ ...blank, ...value });
  const [photo, setPhoto] = useState<HTMLImageElement | null>(null);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const crop = useRef<() => Promise<Blob | null>>(async () => null);
  const set = <K extends keyof typeof draft>(key: K, v: (typeof draft)[K]) =>
    setDraft((d) => ({ ...d, [key]: v }));

  const pick = (file?: File) => {
    if (!file) return;
    if (!/^image\/(png|jpeg|webp)$/.test(file.type)) {
      setError("Choose a PNG, JPEG or WebP image.");
      return;
    }
    const img = new Image();
    img.onload = () => setPhoto(img);
    img.src = URL.createObjectURL(file);
  };

  async function save(e: React.FormEvent) {
    e.preventDefault();
    setBusy(true);
    setError("");
    try {
      let photoUrl = draft.photoUrl;
      if (photo) {
        const blob = await crop.current();
        if (!blob) throw new Error("The photo couldn’t be cropped. Try another image.");
        const form = new FormData();
        form.append("file", new File([blob], "author.jpg", { type: "image/jpeg" }));
        form.append("public", "true");
        const res = await apiFetch("files/upload", { method: "POST", body: form });
        const body = await res.json().catch(() => ({}));
        if (!res.ok || !body.url) throw new Error(body.error || "The photo couldn’t be uploaded. Try again.");
        photoUrl = body.url;
      }
      await request(
        `marketing/authors${value.id ? `/${value.id}` : ""}`,
        value.id ? "PUT" : "POST",
        { ...draft, photoUrl },
      );
      saved();
      toast.success(value.id ? "Author saved" : "Author added");
      close();
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setBusy(false);
    }
  }

  return (
    <Modal
      open
      onOpenChange={close}
      title={value.id ? `Edit ${value.name}` : "Add an author"}
      description="Who readers see an email or issue is from."
    >
      <form className={workspaceClassName("product-form")} onSubmit={save}>
        <div className="flex items-center gap-4">
          {photo ? (
            <SquareCrop image={photo} cropRef={crop} />
          ) : (
            <Avatar author={draft} size={96} />
          )}
          <div className="space-y-2">
            <label className="inline-flex cursor-pointer items-center gap-2 rounded-md border px-3 py-2 text-sm hover:bg-muted">
              <ImagePlus className="size-4" />
              {draft.photoUrl || photo ? "Change photo" : "Upload photo"}
              <input type="file" accept="image/png,image/jpeg,image/webp" className="sr-only" onChange={(e) => pick(e.target.files?.[0])} />
            </label>
            {photo && <p className="text-xs text-muted-foreground">Drag to move the photo and use the slider to zoom.</p>}
            {(draft.photoUrl || photo) && (
              <button
                type="button"
                className="block text-xs text-muted-foreground underline"
                onClick={() => {
                  setPhoto(null);
                  set("photoUrl", "");
                }}
              >
                Remove photo
              </button>
            )}
          </div>
        </div>
        <Field label="Name">
          <input required maxLength={80} value={draft.name} onChange={(e) => set("name", e.target.value)} placeholder="e.g. Asha Rao" />
        </Field>
        <Field label="Email for replies (optional)" hint="Replies go here unless a campaign sets its own reply address.">
          <input type="email" maxLength={254} value={draft.email} onChange={(e) => set("email", e.target.value)} placeholder="asha@yourcompany.com" />
        </Field>
        <fieldset className="space-y-2 text-sm">
          <legend className="mb-1 font-medium">Show the byline</legend>
          <label className="flex items-center gap-2">
            <input type="checkbox" className="!w-4" checked={draft.bylineEmail} onChange={(e) => set("bylineEmail", e.target.checked)} />
            At the top of emails
          </label>
          <label className="flex items-center gap-2">
            <input type="checkbox" className="!w-4" checked={draft.bylineWeb} onChange={(e) => set("bylineWeb", e.target.checked)} />
            Under the issue’s title on your newsletter website
          </label>
          <label className="flex items-center gap-2">
            <input type="checkbox" className="!w-4" checked={draft.isDefault} onChange={(e) => set("isDefault", e.target.checked)} />
            Default author for campaigns and newsletters that don’t pick one
          </label>
        </fieldset>
        {error && (
          <p role="alert" className="text-sm text-destructive">
            {error}
          </p>
        )}
        <div className="flex justify-end gap-2">
          <Button type="button" variant="ghost" onClick={close}>
            Cancel
          </Button>
          <Button type="submit" disabled={busy}>
            {busy ? "Saving…" : "Save author"}
          </Button>
        </div>
      </form>
    </Modal>
  );
}

const VIEW = 96;
const OUT = 256;

/** Draws image covering a size×size square, zoomed and moved (x, y in -1…1). */
function draw(ctx: CanvasRenderingContext2D, image: HTMLImageElement, size: number, zoom: number, x: number, y: number) {
  const scale = Math.max(size / image.naturalWidth, size / image.naturalHeight) * zoom;
  const w = image.naturalWidth * scale;
  const h = image.naturalHeight * scale;
  ctx.clearRect(0, 0, size, size);
  ctx.drawImage(image, (size - w) / 2 + (x * (w - size)) / 2, (size - h) / 2 + (y * (h - size)) / 2, w, h);
  return { w, h };
}

/** A square photo crop: drag to move, slider to zoom. cropRef exports it. */
function SquareCrop({
  image,
  cropRef,
}: {
  image: HTMLImageElement;
  cropRef: React.MutableRefObject<() => Promise<Blob | null>>;
}) {
  const canvas = useRef<HTMLCanvasElement>(null);
  const [zoom, setZoom] = useState(1);
  const [pos, setPos] = useState({ x: 0, y: 0 });
  const drag = useRef<{ px: number; py: number; x: number; y: number } | null>(null);
  useEffect(() => {
    const ctx = canvas.current?.getContext("2d");
    if (ctx) draw(ctx, image, VIEW, zoom, pos.x, pos.y);
    cropRef.current = () => {
      const out = document.createElement("canvas");
      out.width = out.height = OUT;
      const ctx = out.getContext("2d");
      if (!ctx) return Promise.resolve(null);
      ctx.fillStyle = "#fff";
      ctx.fillRect(0, 0, OUT, OUT);
      draw(ctx, image, OUT, zoom, pos.x, pos.y);
      return new Promise((resolve) => out.toBlob(resolve, "image/jpeg", 0.9));
    };
  }, [image, zoom, pos, cropRef]);
  const clamp = (v: number) => Math.max(-1, Math.min(1, v));
  return (
    <div className="space-y-2">
      <canvas
        ref={canvas}
        width={VIEW}
        height={VIEW}
        aria-label="Photo crop"
        className="cursor-move touch-none rounded-full border"
        onPointerDown={(e) => {
          e.currentTarget.setPointerCapture(e.pointerId);
          drag.current = { px: e.clientX, py: e.clientY, ...pos };
        }}
        onPointerMove={(e) => {
          const d = drag.current;
          if (!d) return;
          const scale = Math.max(VIEW / image.naturalWidth, VIEW / image.naturalHeight) * zoom;
          const spareX = image.naturalWidth * scale - VIEW;
          const spareY = image.naturalHeight * scale - VIEW;
          setPos({
            x: spareX > 0 ? clamp(d.x + (2 * (e.clientX - d.px)) / spareX) : 0,
            y: spareY > 0 ? clamp(d.y + (2 * (e.clientY - d.py)) / spareY) : 0,
          });
        }}
        onPointerUp={() => (drag.current = null)}
      />
      <input
        type="range"
        aria-label="Zoom"
        min={1}
        max={3}
        step={0.01}
        value={zoom}
        onChange={(e) => setZoom(Number(e.target.value))}
        className="block w-24"
      />
    </div>
  );
}

/**
 * Picks a campaign's or newsletter's author; "" uses the workspace default.
 * Shows nothing until the workspace has authors.
 */
export function AuthorSelect({
  value,
  onChange,
  className,
  id,
}: {
  value: string;
  onChange: (id: string) => void;
  className?: string;
  id?: string;
}) {
  const q = useMarketingQuery<Author[]>("marketing/authors");
  if (!q.data?.length) return null;
  const fallback = q.data.find((a) => a.isDefault);
  return (
    <select id={id} className={className} value={value} onChange={(e) => onChange(e.target.value)}>
      <option value="">{fallback ? `Default author (${fallback.name})` : "No author"}</option>
      {q.data.map((a) => (
        <option key={a.id} value={a.id}>
          {a.name}
        </option>
      ))}
    </select>
  );
}

/** Whether the workspace has any authors to pick from. */
export function useHasAuthors() {
  return !!useMarketingQuery<Author[]>("marketing/authors").data?.length;
}
