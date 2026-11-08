import type { Metadata, Viewport } from "next";
import Script from "next/script";
import { DM_Sans, EB_Garamond } from "next/font/google";
import { siteUrl } from "@/lib/site";
import "./globals.css";
import { ReferralCapture } from "@/components/referral-capture";
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
    "Open-source email marketing that Claude and other AI agents can run through its MCP server: campaigns, newsletters, automations, forms, a CRM and analytics. Priced per email: free for 3,000 a month, $9 for 20,000, $11.50 for 25,000, or self-host it for free.",
  openGraph: {
    type: "website",
    siteName: "Xem",
    title: "All your email marketing, at the price of email.",
    description:
      "Campaigns, newsletters, automations, forms, CRM and a shared inbox in one open-source suite, priced per email sent. Bring your list from Mailchimp, beehiiv, Kit or Substack.",
    images: [{ url: "/opengraph-image.png", width: 1200, height: 630 }],
  },
  twitter: {
    card: "summary_large_image",
    description:
      "Open-source email marketing with an MCP server for Claude and other AI agents: campaigns, newsletters, automations and a CRM, priced per email sent.",
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
        <ReferralCapture />
        {/* Google tag (gtag.js) */}
        <Script
          src="https://www.googletagmanager.com/gtag/js?id=G-GC39XF1JWS"
          strategy="afterInteractive"
        />
        <Script id="google-tag" strategy="afterInteractive">
          {`window.dataLayer = window.dataLayer || [];
function gtag(){dataLayer.push(arguments);}
gtag('js', new Date());
gtag('config', 'G-GC39XF1JWS');`}
        </Script>
      </body>
    </html>
  );
}
