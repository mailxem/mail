"use client";

import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

import { useEffect } from "react";

export default function Error({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  return (
    <div className="container mx-auto flex flex-col items-center justify-center h-full py-12">
      <div className="mb-6">
        <img
          src="/assets/404.png"
          alt="Page not found illustration"
          width={250}
          className="mx-auto object-contain"
        />
      </div>
      <span className="text-4xl font-medium">Oops! Sorry</span>
      <p className="text-muted-foreground text-center mb-4">
        The page you are looking for doesn't exist or has been moved.
      </p>
      <div className="flex gap-4">
        <Button asChild variant="default">
          <Link href="/">Go Home</Link>
        </Button>
      </div>
    </div>
  );
}
