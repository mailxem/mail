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
  useEffect(() => {
    // Log the error to an error reporting service
    console.error(error);
  }, [error]);
  return (
    <div className="container mx-auto flex flex-col items-center justify-center h-full py-12">
      <div className="mb-6">
        <img
          src="/assets/404.png"
          alt="Page not found illustration"
          width={250}
          height={200}
          className="mx-auto"
        />
      </div>
      <p className="text-muted-foreground text-center mb-6">
        {error.message}
      </p>
      <div className="flex gap-4">
        <Button asChild variant="default">
          <Link href="/">Go Home</Link>
        </Button>
      </div>
    </div>
  );
}
