import "server-only";
import { cache } from "react";
import fs from "node:fs";
import path from "node:path";
import matter from "gray-matter";
import { z } from "zod";
import { unified } from "unified";
import remarkParse from "remark-parse";
import { visit } from "unist-util-visit";
import { toString } from "mdast-util-to-string";
import { remarkHeadingIds } from "./markdown";
import type { PostMeta } from "./blog-types";

const date = z
  .string()
  .regex(/^\d{4}-\d{2}-\d{2}$/)
  .refine((v) => !Number.isNaN(Date.parse(v)), "Invalid date");
const schema = z.object({
  title: z.string().min(15).max(90),
  description: z.string().min(80).max(170),
  date,
  updated: date,
  category: z.enum([
    "AI & writing",
    "Deliverability",
    "Growth",
    "Design",
    "Analytics",
    "Automation",
  ]),
  tags: z.array(z.string().min(2)).min(2).max(6),
  cover: z.string().regex(/^\/images\/blog\/[a-z0-9-]+\.webp$/),
  coverAlt: z.string().min(15),
  coverSource: z.url().startsWith("https://"),
  author: z.literal("Xem editorial"),
  featured: z.boolean().default(false),
  published: z.boolean().default(true),
});
const directory = path.join(process.cwd(), "content/blog");
export const getPosts = cache((): PostMeta[] =>
  fs
    .readdirSync(directory)
    .filter((name) => name.endsWith(".md"))
    .flatMap((name) => {
      const slug = name.slice(0, -3);
      if (!/^[a-z0-9]+(?:-[a-z0-9]+)*$/.test(slug))
        throw new Error(`Invalid blog slug: ${slug}`);
      const { data, content } = matter(
        fs.readFileSync(path.join(directory, name), "utf8"),
      );
      const { published, ...meta } = schema.parse(data);
      if (!published) return [];
      if (meta.updated < meta.date)
        throw new Error(`Update precedes publication: ${slug}`);
      if (!fs.existsSync(path.join(process.cwd(), "public", meta.cover)))
        throw new Error(`Missing cover: ${slug}`);
      return [
        {
          ...meta,
          slug,
          readingMinutes: Math.max(
            1,
            Math.ceil(content.trim().split(/\s+/).length / 220),
          ),
        },
      ];
    })
    .sort(
      (a, b) =>
        b.date.localeCompare(a.date) ||
        Number(b.featured) - Number(a.featured) ||
        a.title.localeCompare(b.title),
    ),
);

export const getPost = cache((slug: string) => {
  const meta = getPosts().find((p) => p.slug === slug);
  if (!meta) return null;
  const { content } = matter(
    fs.readFileSync(path.join(directory, `${meta.slug}.md`), "utf8"),
  );
  const tree = unified().use(remarkParse).parse(content);
  remarkHeadingIds()(tree);
  const toc: { id: string; text: string; depth: number }[] = [];
  visit(tree, "heading", (node) => {
    if (node.depth === 2 || node.depth === 3)
      toc.push({
        id: String(node.data?.hProperties?.id),
        text: toString(node),
        depth: node.depth,
      });
  });
  return { meta, content, toc };
});
export function relatedPosts(post: PostMeta) {
  return getPosts()
    .filter((p) => p.slug !== post.slug)
    .sort(
      (a, b) =>
        Number(b.category === post.category) -
        Number(a.category === post.category),
    )
    .slice(0, 3);
}
