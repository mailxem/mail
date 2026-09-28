import Image from "next/image";
import Link from "next/link";
import { ArrowUpRight } from "lucide-react";
import type { PostMeta } from "@/lib/blog-types";
export function PostCard({ post }: { post: PostMeta }) {
  return (
    <article className="group min-w-0">
      <Link href={`/blog/${post.slug}`} className="block">
        <div className="relative aspect-[3/2] overflow-hidden rounded-2xl bg-lavender">
          <Image
            src={post.cover}
            alt={post.coverAlt}
            fill
            sizes="(max-width:640px) 92vw, (max-width:1024px) 45vw, 370px"
            className="object-cover transition-transform duration-700 group-hover:scale-105 motion-reduce:transform-none"
          />
          <span className="absolute bottom-4 right-4 flex h-10 w-10 items-center justify-center rounded-full bg-cream text-ink transition-colors group-hover:bg-iris group-hover:text-white">
            <ArrowUpRight size={18} />
          </span>
        </div>
        <div className="mt-5 flex items-center gap-3 text-[10px] font-medium uppercase tracking-[0.12em]">
          <span className="text-iris">{post.category}</span>
          <span className="text-muted">{post.readingMinutes} min read</span>
        </div>
        <h3 className="mt-3 font-editorial text-[30px] leading-[1.08] tracking-tight transition-colors group-hover:text-iris">
          {post.title}
        </h3>
        <p className="mt-3 text-sm leading-relaxed text-muted">
          {post.description}
        </p>
      </Link>
    </article>
  );
}
