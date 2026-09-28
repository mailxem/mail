"use client";

import { useEffect, useRef, useState } from "react";
import { ArrowUpRight, Pause, Play } from "lucide-react";
import { SiteDialog } from "./site-dialog";

const film = "/videos/xem-brand-film.mp4";
const preview = "/videos/xem-brand-preview.mp4";
const poster = "/videos/xem-brand-poster.webp";

type DataConnection = EventTarget & { saveData?: boolean };

function playPreview(video: HTMLVideoElement) {
  if (!video.getAttribute("src")) video.src = preview;
  void video.play().catch(() => {
    // Autoplay can be declined by the browser; the poster and play link remain.
  });
}

function FilmPlayer() {
  const [source, setSource] = useState<string>();
  const [failed, setFailed] = useState(false);

  useEffect(() => {
    const controller = new AbortController();
    let objectUrl: string | undefined;

    // Static asset hosts may omit byte-range support. A local object URL lets
    // native controls seek through the complete 2.35 MB film on every host.
    async function loadFilm() {
      try {
        const response = await fetch(film, { signal: controller.signal });
        if (!response.ok) throw new Error("Film unavailable");
        const blob = await response.blob();
        if (controller.signal.aborted) return;
        objectUrl = URL.createObjectURL(blob);
        setSource(objectUrl);
      } catch {
        if (!controller.signal.aborted) setFailed(true);
      }
    }
    void loadFilm();
    return () => {
      controller.abort();
      if (objectUrl) URL.revokeObjectURL(objectUrl);
    };
  }, []);

  return (
    <div className="relative bg-ink" aria-busy={!source && !failed}>
      <video
        src={source}
        poster={poster}
        width={1280}
        height={720}
        autoPlay
        controls
        playsInline
        preload="none"
        tabIndex={0}
        aria-label="Xem brand film"
        aria-describedby="film-description"
        className="aspect-video w-full"
      >
        <track
          kind="captions"
          src="/videos/xem-brand-film.vtt"
          srcLang="en"
          label="English"
        />
        <a href={film}>Watch the Xem brand film</a>
      </video>
      {!source && (
        <p
          role="status"
          className="absolute inset-x-4 top-4 mx-auto w-fit rounded-lg bg-cream px-4 py-3 text-sm text-ink shadow-md"
        >
          {failed ? (
            <>
              The film couldn’t load.{" "}
              <a href={film} className="underline underline-offset-4">
                Open the video directly.
              </a>
            </>
          ) : (
            "Opening the film…"
          )}
        </p>
      )}
    </div>
  );
}

export function HeroFilm({ suspended = false }: { suspended?: boolean }) {
  const card = useRef<HTMLDivElement>(null);
  const video = useRef<HTMLVideoElement>(null);
  const [open, setOpen] = useState(false);
  const [inView, setInView] = useState(false);
  const [pageVisible, setPageVisible] = useState(false);
  const [autoPreview, setAutoPreview] = useState<boolean | null>(null);
  const [previewIntent, setPreviewIntent] = useState<boolean | null>(null);
  const [playing, setPlaying] = useState(false);

  useEffect(() => {
    const motion = window.matchMedia("(prefers-reduced-motion: reduce)");
    const connection = (
      navigator as Navigator & { connection?: DataConnection }
    ).connection;
    const preferences = () => {
      setAutoPreview(!motion.matches && !connection?.saveData);
      setPreviewIntent(null);
    };
    const visibility = () => setPageVisible(!document.hidden);
    preferences();
    visibility();
    motion.addEventListener("change", preferences);
    connection?.addEventListener("change", preferences);
    document.addEventListener("visibilitychange", visibility);

    const observer = new IntersectionObserver(
      ([entry]) =>
        setInView(entry.isIntersecting && entry.intersectionRatio >= 0.15),
      { threshold: 0.15 },
    );
    if (card.current) observer.observe(card.current);

    return () => {
      observer.disconnect();
      motion.removeEventListener("change", preferences);
      connection?.removeEventListener("change", preferences);
      document.removeEventListener("visibilitychange", visibility);
    };
  }, []);

  const shouldPlay =
    (previewIntent ?? autoPreview) &&
    inView &&
    pageVisible &&
    !open &&
    !suspended;

  useEffect(() => {
    const el = video.current;
    if (!el) return;
    if (shouldPlay) playPreview(el);
    else el.pause();
    return () => el.pause();
  }, [shouldPlay]);

  return (
    <>
      <div className="relative mx-auto mb-10 mt-8 max-w-[1120px] px-5 md:mb-14 md:mt-10 md:px-10">
        <p
          aria-hidden="true"
          className="absolute -left-5 top-20 hidden -rotate-[9deg] font-editorial text-[25px] italic leading-[1.05] text-forest xl:block"
        >
          A little more Xem.
          <br />
          In thirty seconds.
          <svg viewBox="0 0 100 90" className="ml-8 mt-3 h-20 w-24" fill="none">
            <path
              d="M14 6C-8 66 63 92 88 37m-20 7 21-10 5 23"
              stroke="currentColor"
              strokeWidth="1.3"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
          </svg>
        </p>

        <div ref={card} className="relative mx-auto max-w-[800px]">
          <div
            aria-hidden="true"
            className="pointer-events-none absolute inset-0 rotate-[-2deg] rounded-xl border border-ink/15 bg-lemon md:rotate-[-3deg]"
          />
          <div className="relative rounded-xl border border-ink/20 bg-[#fffef8] p-2 shadow-[0_22px_55px_-30px_#22251f55] md:p-3">
            <a
              href={film}
              aria-label="Play the Xem brand film, 30 seconds"
              aria-haspopup="dialog"
              onClick={(event) => {
                if (
                  event.button !== 0 ||
                  event.metaKey ||
                  event.ctrlKey ||
                  event.shiftKey ||
                  event.altKey
                )
                  return;
                event.preventDefault();
                setOpen(true);
              }}
              className="group relative block aspect-video overflow-hidden rounded-[5px] bg-iris focus-visible:outline focus-visible:outline-[3px] focus-visible:outline-offset-4 focus-visible:outline-iris"
            >
              <video
                ref={video}
                poster={poster}
                width={1280}
                height={720}
                muted
                loop
                playsInline
                preload="none"
                aria-hidden="true"
                tabIndex={-1}
                onPlaying={() => setPlaying(true)}
                onPause={() => setPlaying(false)}
                className="pointer-events-none h-full w-full object-cover"
              />
              <span className="absolute bottom-2 left-2 inline-flex items-center gap-1.5 rounded-full border border-cream/30 bg-cream px-2 py-1.5 text-[10px] font-medium text-ink shadow-sm transition-colors group-hover:bg-lemon md:bottom-5 md:left-5 md:gap-3 md:px-4 md:py-3 md:text-sm">
                <span className="flex h-5 w-5 items-center justify-center rounded-full bg-ink text-cream md:h-8 md:w-8">
                  <Play size={12} fill="currentColor" aria-hidden="true" />
                </span>
                Play the film
                <span className="ml-1 border-l border-ink/20 pl-2 font-mono text-[10px] tabular-nums md:pl-3 md:text-xs">
                  00:30
                </span>
              </span>
            </a>

            <div className="flex min-h-12 items-center justify-between gap-2 px-2 pt-2 md:min-h-14 md:px-3">
              <p className="font-editorial text-lg italic tracking-tight md:text-[23px]">
                Good things start with a connection.
              </p>
              {autoPreview !== null && (
                <button
                  type="button"
                  aria-label={
                    playing ? "Pause film preview" : "Play film preview"
                  }
                  onClick={() => {
                    setPreviewIntent(!playing);
                    if (!video.current) return;
                    if (playing) video.current.pause();
                    else playPreview(video.current);
                  }}
                  className="flex shrink-0 items-center gap-2 rounded-full p-3 text-muted transition-colors hover:bg-lemon hover:text-ink focus-visible:outline focus-visible:outline-2 focus-visible:outline-iris"
                >
                  {playing ? (
                    <Pause size={13} aria-hidden="true" />
                  ) : (
                    <Play size={13} aria-hidden="true" />
                  )}
                  <span className="hidden text-[10px] uppercase tracking-[.1em] sm:inline">
                    {playing ? "Preview on" : "Preview paused"}
                  </span>
                </button>
              )}
            </div>
          </div>
          <span
            aria-hidden="true"
            className="pointer-events-none absolute -right-7 -top-7 hidden h-[82px] w-[82px] rotate-[12deg] items-center justify-center rounded-full border border-dashed border-forest/60 bg-lemon p-3 text-center text-[10px] font-medium uppercase leading-relaxed tracking-[.12em] shadow-[0_0_0_5px_#edf09b] md:flex"
          >
            A little
            <br />
            more human
          </span>
        </div>
      </div>

      {open && (
        <SiteDialog
          title="A little more human. A film by Xem."
          onClose={() => setOpen(false)}
        >
          <FilmPlayer />
          <div className="flex flex-wrap items-start justify-between gap-4 px-5 py-5 text-sm md:px-7">
            <details className="max-w-2xl">
              <summary
                tabIndex={0}
                className="cursor-pointer font-medium underline decoration-ink/30 underline-offset-4"
              >
                About the film &amp; text description
              </summary>
              <p
                id="film-description"
                className="mt-3 leading-relaxed text-muted"
              >
                A 30-second journey through Xem’s world of open-source email.
                Our connected mark becomes colors, letters, newsletters, and a
                message traveling through an illustrated landscape. The film
                introduces design, sending, automation, and audience insights,
                then closes with “Make email more human.” Instrumental music; no
                speech.
              </p>
            </details>
            <a
              href={film}
              download
              className="inline-flex items-center gap-1.5 font-medium underline decoration-ink/30 underline-offset-4 hover:text-iris"
            >
              Download film <ArrowUpRight size={15} aria-hidden="true" />
            </a>
          </div>
        </SiteDialog>
      )}
    </>
  );
}
