import type { Metadata, Viewport } from "next";
import { DM_Sans, EB_Garamond } from "next/font/google";
import { siteUrl } from "@/lib/site";
import "./globals.css";
const sans = DM_Sans({
  subsets: ["latin"],
  variable: "--font-sans",
  display: "swap",
});
const editorial = EB_Garamond({
  subsets: ["latin"],
  variable: "--font-editorial",
  style: ["normal", "italic"],
  display: "swap",
});
export const metadata: Metadata = {
  metadataBase: new URL(siteUrl),
  title: "Xem — All your email marketing, at the price of email.",
  description:
    "The open-source email marketing suite: campaigns, newsletters, automations, forms, a CRM, a shared inbox and analytics. Send through Amazon SES and pay what Amazon charges: $2.00 to $3.20 for 20,000 emails.",
  openGraph: {
    type: "website",
    siteName: "Xem",
    title: "All your email marketing, at the price of email.",
    description:
      "Campaigns, newsletters, automations, forms, CRM and a shared inbox in one open-source suite, sent through Amazon SES at Amazon's price. Bring your list from Mailchimp, beehiiv, Kit or Substack.",
    images: [{ url: "/opengraph-image.png", width: 1200, height: 630 }],
  },
  twitter: {
    card: "summary_large_image",
    description:
      "The open-source email marketing suite: campaigns, newsletters, automations, CRM and a shared inbox, at Amazon SES prices.",
    title: "Xem — All your email marketing, at the price of email.",
    images: ["/opengraph-image.png"],
  },
  icons: { icon: "/favicon.ico", apple: "/brand/xem-mark.png" },
};
export const viewport: Viewport = { themeColor: "#ffffef" };
export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html
      lang="en"
      className={`${sans.variable} ${editorial.variable} scroll-smooth motion-reduce:scroll-auto`}
    >
      <body className="bg-cream font-sans text-ink antialiased selection:bg-lavender selection:text-ink [&_a]:outline-offset-4 [&_button]:outline-offset-4 [&_:focus-visible]:outline-2 [&_:focus-visible]:outline-iris">
        <a
          href="#main"
          className="fixed left-4 top-4 z-[100] -translate-y-24 rounded-lg bg-ink px-5 py-3 text-cream focus:translate-y-0"
        >
          Skip to content
        </a>
        {children}
      </body>
    </html>
  );
}
