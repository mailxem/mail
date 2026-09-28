import { workspaceClassName } from "@/lib/workspace-styles";
import "./globals.css";
import { ConfirmSheetProvider } from "@/components/ui/confirm-sheet";
import { AppShell } from "@/components/app-shell";
import type { Metadata } from "next";
import { ThemeProvider } from "@/components/theme-provider";
import { auth } from "@/auth";
import { NextAuthProvider } from "./providers/NextAuthProvider";
import { TeamProvider } from "@/app/providers/team-provider";
import { Toaster } from "sonner";
import { cn } from "@/lib/utils";
import { QueryProvider } from "./providers/query-provider";
import { GeistSans } from "geist/font/sans";
import { GeistMono } from "geist/font/mono";

export const metadata: Metadata = {
  keywords: [
    "email",
    "automation",
    "email marketing",
    "email management",
    "email templates",
    "email campaigns",
  ],
  authors: [{ name: "Harsh Vardhan Goswami" }],
  robots: "index, follow",
};

async function MainLayout({ children }: { children: React.ReactNode }) {
  const session = await auth();

  if (!session) {
    return <>{children}</>;
  }

  return <AppShell>{children}</AppShell>;
}

export default async function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en" suppressHydrationWarning>
      <head>
        <title>Xem · Email, newsletters & customer journeys</title>
        <meta
          name="description"
          content="Create newsletters, automate customer journeys, and grow your audience with Xem."
        />

        <meta property="og:url" content="https://xem.email" />
        <meta property="og:type" content="website" />
        <meta
          property="og:title"
          content="Xem · Email, newsletters & customer journeys"
        />
        <meta
          property="og:description"
          content="Create newsletters, automate customer journeys, and grow your audience with Xem."
        />
        <meta
          property="og:image"
          content="https://framerusercontent.com/images/BIW4segud1fRHZGJB3wIGXKg9DQ.png"
        />

        {/* Twitter Meta Tags */}
        <meta name="twitter:card" content="summary_large_image" />
        <meta property="twitter:domain" content="xem.email" />
        <meta property="twitter:url" content="https://xem.email" />
        <meta
          name="twitter:title"
          content="Xem · Email, newsletters & customer journeys"
        />
        <meta
          name="twitter:description"
          content="Create newsletters, automate customer journeys, and grow your audience with Xem."
        />
        <meta
          name="format-detection"
          content="telephone=no, date=no, email=no, address=no"
        />
        <link
          href="https://api.fontshare.com/v2/css?f[]=azeret-mono@400,500,600,700&display=swap"
          rel="stylesheet"
        />
        <link rel="preconnect" href="https://fonts.googleapis.com" />
        <link
          rel="preconnect"
          href="https://fonts.gstatic.com"
          crossOrigin="anonymous"
        />
        {/* publoic manifest */}
        <link rel="manifest" href="/site.webmanifest" />
        {/* publoic icons */}
        <link
          rel="apple-touch-icon"
          sizes="180x180"
          href="/apple-touch-icon.png"
        />
        <link
          rel="icon"
          type="image/png"
          sizes="32x32"
          href="/favicon-32x32.png"
        />
        <link
          rel="icon"
          type="image/png"
          sizes="16x16"
          href="/favicon-16x16.png"
        />
        {/* favicon */}
        <link rel="icon" type="image/png" sizes="32x32" href="/favicon.ico" />
        {/* theme color */}
        <meta name="theme-color" content="#191919" />
      </head>
      <body
        className={workspaceClassName(cn(
          "document-theme antialiased bg-background text-foreground font-sans text-sm",
          GeistSans.variable,
          GeistMono.variable,
        ))}
      >
        <ThemeProvider
          attribute="class"
          defaultTheme="dark"
          storageKey="xem-appearance"
          enableSystem
        >
          <NextAuthProvider>
            <QueryProvider>
              <TeamProvider>
                <ConfirmSheetProvider>
                  <MainLayout>{children}</MainLayout>
                </ConfirmSheetProvider>
              </TeamProvider>
            </QueryProvider>
          </NextAuthProvider>
        </ThemeProvider>
        <Toaster />
      </body>
    </html>
  );
}
