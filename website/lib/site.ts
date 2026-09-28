function origin(value: string | undefined, fallback: string) {
  const url = new URL(value || fallback);
  if (!["http:", "https:"].includes(url.protocol))
    throw new Error("Site URLs must use HTTP or HTTPS.");
  return url.origin;
}
export const siteUrl = origin(
  process.env.NEXT_PUBLIC_SITE_URL,
  "https://xem.email",
);
export const appUrl = origin(
  process.env.NEXT_PUBLIC_APP_URL,
  "https://app.xem.email",
);
export const appLink = (path: string) => `${appUrl}${path}`;

export const githubUrl = "https://github.com/mailxem/mail";
export const selfHostUrl = `${githubUrl}/blob/undefined/devops/swarm/README.md`;
export const contributeUrl = `${githubUrl}#contribute`;
