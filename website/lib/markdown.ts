import GithubSlugger from "github-slugger";
import { toString } from "mdast-util-to-string";
import { visit } from "unist-util-visit";
import type { Root } from "mdast";

export function remarkHeadingIds() {
  return (tree: Root) => {
    const slugger = new GithubSlugger();
    visit(tree, "heading", (node) => {
      node.data ??= {};
      node.data.hProperties = {
        ...node.data.hProperties,
        id: `section-${slugger.slug(toString(node))}`,
      };
    });
  };
}
