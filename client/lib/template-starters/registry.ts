import manifest from "./manifest.json";

export type TemplateStarter = (typeof manifest)[number];
export const templateStarters: readonly TemplateStarter[] = [...manifest].sort((a, b) =>
  Number(b.collection === "Reference designs") - Number(a.collection === "Reference designs"),
);
export const starterCategories = [
  "All templates",
  "Newsletters",
  "Welcome",
  "Product",
  "Commerce",
  "Events",
  "Transactional",
  "Retention",
] as const;

export function getTemplateStarter(key: string | null | undefined) {
  return templateStarters.find((starter) => starter.key === key);
}

export const starterCollections = ["All collections", "Reference designs", "Xem originals", "Editorial collection"] as const;

export function filterTemplateStarters(category: string, search: string, collection = "All collections") {
  const term = search.trim().toLowerCase();
  return templateStarters.filter(
    (starter) =>
      (category === "All templates" || starter.category === category) &&
      (collection === "All collections" || starter.collection === collection) &&
      [starter.name, starter.description, starter.category, ...starter.tags]
        .join(" ")
        .toLowerCase()
        .includes(term),
  );
}

export async function loadStarterDesign(key: string, signal?: AbortSignal) {
  const starter = getTemplateStarter(key);
  if (!starter)
    throw new Error(
      "This starter template could not be found. Choose another from the template library.",
    );
  const response = await fetch(starter.designUrl, { signal });
  if (!response.ok)
    throw new Error(
      "The starter design could not be loaded. Please try again.",
    );
  const design = await response.json();
  if (
    !Array.isArray(design?.body?.rows) ||
    !design.body.rows.length ||
    !design.schemaVersion
  ) {
    throw new Error(
      "This starter design is invalid. Please choose another template.",
    );
  }
  return resolveStarterAssets(design, window.location.origin);
}

// Unlayer runs in a cross-origin iframe, and sent emails need absolute image URLs.
// Only resolve our registry-owned asset paths; never rewrite user template URLs.
export function resolveStarterAssets<T>(design: T, origin: string): T {
  return JSON.parse(JSON.stringify(design), (_key, value) =>
    typeof value === "string" && value.startsWith("/assets/template-starters/")
      ? new URL(value, origin).href
      : value,
  );
}
