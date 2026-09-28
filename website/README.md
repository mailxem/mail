# Xem website

The independent Next.js marketing site and Markdown journal for Xem. It exports to plain HTML, CSS, JavaScript, and local images in `out/`, served by **Cloudflare Workers Static Assets**. There is no Next.js server, Worker application code, database, or image-optimization endpoint in production.

## Development

```sh
bun install --frozen-lockfile
cp .env.example .env.local
bun dev
```

Local development runs at http://localhost:3001. `.env.local` can point `NEXT_PUBLIC_SITE_URL` at localhost:3001 and `NEXT_PUBLIC_APP_URL` at localhost:3000. These local overrides are not used by the deployment build.

## Static build and preview

```sh
bun run build
bun run preview
```

`build` exports every route, including all published articles, sitemap, robots, and RSS feed, into `out/`. `preview` serves that output with Wrangler at http://localhost:3002. Use `bunx wrangler dev --port 3003` to select another preview port.

The deployment build explicitly defaults to `https://xem.email` and `https://app.xem.email`, preventing an ignored `.env.local` file from publishing localhost canonicals or signup links. For a different deployment origin, set these **build-time** environment variables:

```sh
XEM_BUILD_SITE_URL=https://www.example.com \
XEM_BUILD_APP_URL=https://app.example.com bun run build
```

Use HTTPS origins. Changing the origin requires rebuilding; Wrangler runtime variables cannot change already exported HTML or browser bundles.

## Deploy with Cloudflare Wrangler

```sh
# First-time authentication on your machine
bunx wrangler login

# Validate build and deployment configuration without uploading
bun run deploy:check

# Build and upload the static website
bun run deploy
```

`wrangler.jsonc` defines the `xem-website` deployment, compatibility date, `out/` asset directory, extensionless URLs without trailing slashes, and the exported `404.html` fallback. It deliberately does not use SPA fallback: unknown blog URLs return 404.

`public/_headers` is copied into the export and supplies security headers, immutable caching for hashed Next.js assets, and RSS content type. The other files use Cloudflare's default revalidation behavior. Images are already optimized local WebP files and are served directly; no paid image transformation service is required.

For noninteractive deployment, supply `CLOUDFLARE_API_TOKEN` (with the required Workers deployment permissions) and `CLOUDFLARE_ACCOUNT_ID` through your CI secret store. Do not commit tokens, `.dev.vars`, or account credentials.

### Cloudflare Workers Builds / Git integration

Connect this repository in Cloudflare **Workers & Pages → Create application → Import a repository** and choose Workers:

- Root directory: repository root
- Install command: `bun install --frozen-lockfile`
- Build command: `bun run build`
- Deploy command: `bunx wrangler deploy`
- Worker name: `xem-website`
- Build environment: Bun 1.3.5 (or compatible), Node.js 22+; optionally set the two `XEM_BUILD_*` variables above.

After the first deployment, add the intended production hostname under **Settings → Domains & Routes**. This configuration does not change the existing live `xem.email` domain or replace the old website automatically. Verify the new deployment before switching the domain. Submit the production `/sitemap.xml` in Search Console when the domain points to the new site.

## Verification

```sh
bun run typecheck
bun run deploy:check
bun run preview
# In another terminal:
bun run test:e2e
```

Tests use installed Google Chrome. Set `PLAYWRIGHT_BASE_URL` to use another preview URL. Coverage includes interactive analytics, exports, template previews, keyboard navigation, mobile layouts, accessibility, Markdown articles, metadata, sitemap/RSS, images, and no-JavaScript content.

## Content and design

- `components/hero-film.tsx`: the hero's paper-framed film preview and accessible full player. The 489 KB silent preview loads when visible, pauses offscreen or while a dialog is open, and respects reduced motion and data saver. The full film loads only on request. The poster and direct film link also work without JavaScript.
- `public/videos/`: optimized 720p film with audio, 832px silent preview, WebP poster, and music captions. The full player fetches the 2.35 MB film after opening and uses a temporary object URL so native seeking also works on static hosts without byte-range support; closing aborts the fetch and releases the URL. Editable source, the 1080p master, reference notes, and render verification live in [`videos/xem-brand-film/`](videos/xem-brand-film/README.md).
- `content/blog/*.md`: published articles. Set `published: false` for a draft. Frontmatter, dates, bylines, cover paths, and metadata are validated at build time. Rebuild after changing content.
- `lib/blog.ts`: build-time Markdown loading; raw HTML is skipped during rendering. Article routes, related stories, reading times, and contents are generated from the files.
- `scripts/render-managed-smtp-art.mjs`: renders the original managed SMTP cover and three explanatory SVG diagrams. Run with `node scripts/render-managed-smtp-art.mjs`; the illustrations contain no measured performance data.
- `lib/brand.ts`: shared iris (`#5b3cc4`) primary buttons.
- `lib/content.ts`: product sections and FAQs.
- `lib/analytics-demo.ts`: deterministic, clearly labeled synthetic homepage analytics. No customer data is fetched.
- `components/sending-section.tsx`: illustrative, local-only preview of the upcoming setup checklist. Keep managed sending labeled as coming soon until the hosted runtime and live feedback checks are complete; SES production approval alone is not a customer launch. The provider connection CTA uses the existing settings page.
- `lib/templates.json`: the first 15 starters in the product gallery's order.
- `public/images/`: local product captures, template previews, and blog photographs.
- `design/research/`: source notes, qualitative keyword research, photo provenance, and original testimonial attribution.
- `design/reference/` and `design/review/`: design references and visual verification captures; not included in the static export.

The website uses Tailwind, GSAP, Three.js, and shadcn/Recharts. All marketing interactions stay in the browser; signup links open the separate Xem product. Product screenshots show illustrative workspace data. Template brand names are part of reference designs, not customer endorsements. The journal explicitly describes its AI-assisted editorial process at `/blog/editorial`.

After replacing the film master, run `node scripts/prepare-hero-film.mjs` (requires FFmpeg) to regenerate the website copies, then rebuild the site. Review the poster timestamp and captions if the film's timing changes.

To refresh template captures with the product running locally:

```sh
bun run capture:templates
```

## Optional container

The Dockerfile builds the same static export and serves it using unprivileged Nginx on port 8080:

```sh
docker build -t xem-website .
docker run --rm -p 3001:8080 xem-website
```

Cloudflare deployments do not require Docker.

The Ask Xem homepage showcase uses `@shadcn/helpers/ai-sdk` with the real `useChat` lifecycle to animate a deterministic example. It never calls the inference proxy or workspace APIs. Play/pause/replay controls and reduced-motion rendering are included. Keep its development label until the assistant runtime is deployed.
