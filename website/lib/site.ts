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
export const apiUrl = origin(
  process.env.NEXT_PUBLIC_API_URL,
  "https://api.xem.email",
);
export const appLink = (path: string) => `${appUrl}${path}`;

export const githubUrl = "https://github.com/mailxem/mail";
export const docsUrl = "https://docs.xem.email";
export const selfHostUrl = `${githubUrl}/blob/undefined/devops/swarm/README.md`;
// Dodo Payments checkout for the one-time $9 self-hosted license; Dodo emails the key.
export const licenseUrl =
  "https://checkout.dodopayments.com/buy/pdt_0NrPhxW0twqMx4Oyk3BCU?quantity=1";
export const contributeUrl = `${githubUrl}#contribute`;
