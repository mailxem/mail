"use client";

import * as React from "react";

export function useMediaQuery(query: string) {
  const [matches, setMatches] = React.useState(false);

  React.useEffect(() => {
    const media = window.matchMedia(query);
    
    // Set initial value
    setMatches(media.matches);

    // Create listener function
    const listener = (e: MediaQueryListEvent) => {
      setMatches(e.matches);
    };

    // Add listener
    media.addEventListener("change", listener);

    // Clean up
    return () => {
      media.removeEventListener("change", listener);
    };
  }, [query]);

  return matches;
}
