"use client";
import { Dither, skyField } from "@/components/illustrations/dither";

/** Dithered blue sky used on the Get started and Upgrade cards. */
export function GridIllustration({ className }: { className?: string }) {
  return <Dither field={skyField} ink="#2f63c9" inkDark="#6f95f0" className={className} />;
}
