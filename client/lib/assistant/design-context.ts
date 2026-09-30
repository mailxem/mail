// Keep the actual email content; remove export scaffolding before sending it to
// the model. Never silently truncate a customer's long template.
export function designContext(html: string): string {
  const document = new DOMParser().parseFromString(html, "text/html");
  document
    .querySelectorAll("script,style,meta,link")
    .forEach((node) => node.remove());
  const walker = document.createTreeWalker(
    document.body,
    NodeFilter.SHOW_COMMENT,
  );
  const comments: Node[] = [];
  while (walker.nextNode()) comments.push(walker.currentNode);
  comments.forEach((node) => node.parentNode?.removeChild(node));
  const content = document.body.innerHTML.trim();
  if (content.length > 16_000)
    throw new Error(
      "This email is too long for a complete AI revision. Shorten it in the editor first so Xem can preserve all of your content.",
    );
  return content;
}
