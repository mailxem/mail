import { spawnSync } from "node:child_process";

// Production builds must not inherit localhost URLs from Bun's .env.local loading.
// Explicit deployment overrides use separate build-only variables.
const site = new URL(process.env.XEM_BUILD_SITE_URL || "https://xem.email");
const app = new URL(process.env.XEM_BUILD_APP_URL || "https://app.xem.email");
for (const url of [site, app]) {
  if (url.protocol !== "https:")
    throw new Error("Static deployment URLs must use HTTPS.");
}
const result = spawnSync(
  process.execPath,
  ["node_modules/next/dist/bin/next", "build"],
  {
    stdio: "inherit",
    env: {
      ...process.env,
      NEXT_PUBLIC_SITE_URL: site.origin,
      NEXT_PUBLIC_APP_URL: app.origin,
      NEXT_TELEMETRY_DISABLED: "1",
    },
  },
);
if (result.error) throw result.error;
process.exit(result.status ?? 1);
