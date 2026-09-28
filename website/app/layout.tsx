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
  title: "Xem — Every email, a little more human.",
  description:
    "Open-source email marketing with Xem. Design beautiful emails, send newsletters, and build customer journeys. Explore the code and make it your own.",
  openGraph: {
    type: "website",
    siteName: "Xem",
    title: "Every email, a little more human.",
    description:
      "Proudly open-source email marketing. Beautiful emails, thoughtful automations, and room to make it your own.",
    images: [{ url: "/opengraph-image.png", width: 1200, height: 630 }],
  },
  twitter: {
    card: "summary_large_image",
    description:
      "Proudly open-source email marketing. Explore Xem, build on it, and help shape what comes next.",
    title: "Xem — Every email, a little more human.",
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
