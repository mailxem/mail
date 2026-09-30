// Generated text blocks can contain the brand lockup as inline HTML. Keep the
// markup intact and change only quoted image sources, never links or prose.
export function mapInlineImageSources(
  html: string,
  map: (url: string) => string,
): string {
  return html.replace(
    /(<img\b[^>]*?\ssrc\s*=\s*)(["'])([^"'<>]*)\2/gi,
    (_match, prefix, quote, url) => `${prefix}${quote}${map(url)}${quote}`,
  );
}
