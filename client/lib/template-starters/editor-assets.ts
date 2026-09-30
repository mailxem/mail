import { mapInlineImageSources } from "./inline-images";

type AssetSourceMap = Map<string, string>;
const localHost = (hostname: string) =>
  ["localhost", "127.0.0.1", "[::1]"].includes(hostname);

// A hosted editor cannot request Chrome's loopback address space. On local
// development only, supply our own raster assets as data URLs through loadDesign.
// Export reverses this preview-only adaptation so saved mail uses hosted URLs.
export async function prepareEditorAssets<T>(
  design: T,
  signal: AbortSignal,
  origin = window.location.origin,
): Promise<{ design: T; sources: AssetSourceMap }> {
  const sources: AssetSourceMap = new Map();
  if (!localHost(new URL(origin).hostname))
    return { design: structuredClone(design), sources };
  const urls = new Set<string>();
  const collect = (value: string) => {
    if (
      value.startsWith(`${origin}/assets/template-starters/`) &&
      /\.(?:png|jpe?g|webp|gif|svg)(?:[?#].*)?$/i.test(value)
    )
      urls.add(value);
    return value;
  };
  const walk = (value: unknown): void => {
    if (typeof value === "string")
      mapInlineImageSources(collect(value), collect);
    else if (Array.isArray(value)) value.forEach(walk);
    else if (value && typeof value === "object")
      Object.values(value).forEach(walk);
  };
  walk(design);
  const replacements = new Map<string, string>();
  await Promise.all(
    [...urls].map(async (url) => {
      const response = await fetch(url, { signal });
      if (!response.ok)
        throw new Error("A template image could not be loaded. Please retry.");
      const type = response.headers.get("content-type")?.split(";")[0] ?? "";
      if (
        !["image/png", "image/jpeg", "image/webp", "image/gif"].includes(type)
      )
        throw new Error("Unsupported template image format.");
      const bytes = new Uint8Array(await response.arrayBuffer());
      if (bytes.length > 5 * 1024 * 1024)
        throw new Error("This template image is too large.");
      let binary = "";
      for (let offset = 0; offset < bytes.length; offset += 8192)
        binary += String.fromCharCode(...bytes.subarray(offset, offset + 8192));
      const dataURL = `data:${type};base64,${btoa(binary)}`;
      replacements.set(url, dataURL);
      sources.set(dataURL, url);
    }),
  );
  const replace = (value: string) => replacements.get(value) ?? value;
  return {
    design: JSON.parse(JSON.stringify(design), (_key, value) =>
      typeof value === "string"
        ? mapInlineImageSources(replace(value), replace)
        : value,
    ),
    sources,
  };
}

export function restoreEditorAssets<T>(
  output: { design: T; html: string },
  sources: AssetSourceMap,
) {
  let html = output.html;
  for (const [dataURL, url] of sources) html = html.replaceAll(dataURL, url);
  const restore = (value: string) => sources.get(value) ?? value;
  return {
    html,
    design: JSON.parse(JSON.stringify(output.design), (_key, value) =>
      typeof value === "string"
        ? mapInlineImageSources(restore(value), restore)
        : value,
    ) as T,
  };
}
