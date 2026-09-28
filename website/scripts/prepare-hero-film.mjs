import { execFileSync } from "node:child_process";
import { mkdirSync } from "node:fs";
import { fileURLToPath } from "node:url";

const master = fileURLToPath(
  new URL(
    "../videos/xem-brand-film/renders/xem-brand-film-1080p.mp4",
    import.meta.url,
  ),
);
const output = new URL("../public/videos/", import.meta.url);
mkdirSync(output, { recursive: true });

function ffmpeg(args, name) {
  execFileSync(
    "ffmpeg",
    [
      "-hide_banner",
      "-loglevel",
      "warning",
      ...args,
      "-y",
      fileURLToPath(new URL(name, output)),
    ],
    { stdio: "inherit" },
  );
}

ffmpeg(
  [
    "-i",
    master,
    "-map",
    "0:v:0",
    "-map",
    "0:a:0",
    "-vf",
    "scale=1280:720",
    "-c:v",
    "libx264",
    "-preset",
    "slow",
    "-crf",
    "23",
    "-pix_fmt",
    "yuv420p",
    "-c:a",
    "aac",
    "-b:a",
    "128k",
    "-movflags",
    "+faststart",
    "-threads",
    "4",
  ],
  "xem-brand-film.mp4",
);

ffmpeg(
  [
    "-i",
    master,
    "-an",
    "-vf",
    "scale=832:468,fps=24",
    "-c:v",
    "libx264",
    "-preset",
    "slow",
    "-crf",
    "29",
    "-pix_fmt",
    "yuv420p",
    "-movflags",
    "+faststart",
    "-threads",
    "4",
  ],
  "xem-brand-preview.mp4",
);

ffmpeg(
  [
    "-ss",
    "26.6",
    "-i",
    master,
    "-frames:v",
    "1",
    "-vf",
    "scale=1280:720",
    "-c:v",
    "libwebp",
    "-quality",
    "85",
  ],
  "xem-brand-poster.webp",
);

console.log(
  "Prepared the full film, silent preview, and poster in public/videos.",
);
