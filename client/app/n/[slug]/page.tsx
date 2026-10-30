import type { Metadata } from "next";
import Link from "next/link";
import { archiveFetch, issueDate, type Archive } from "@/lib/newsletters/archive";
import styles from "../archive.module.css";

type Params = { params: Promise<{ slug: string }> };

export async function generateMetadata({ params }: Params): Promise<Metadata> {
  const { slug } = await params;
  const a = await archiveFetch<Archive>(encodeURIComponent(slug));
  return {
    title: a.name,
    description: a.description || `Every issue of ${a.name}.`,
    openGraph: { title: a.name, description: a.description, type: "website" },
    alternates: {
      types: { "application/rss+xml": `/n/${encodeURIComponent(a.slug)}/feed.xml` },
    },
    robots: "index, follow",
  };
}

export default async function NewsletterArchive({ params }: Params) {
  const { slug } = await params;
  const a = await archiveFetch<Archive>(encodeURIComponent(slug));
  return (
    <main className={styles.page}>
      <header className={styles.masthead}>
        <h1 className={styles.name}>{a.name}</h1>
        {a.description && <p className={styles.lede}>{a.description}</p>}
        {a.subscribeForm && (
          <Link href={`/f/${a.subscribeForm}`} className={styles.subscribe}>
            Subscribe
          </Link>
        )}
      </header>
      {a.issues.length ? (
        <ol className={styles.issues}>
          {a.issues.map((i) => (
            <li key={i.id}>
              <Link href={`/n/${a.slug}/${i.id}`}>
                <time dateTime={i.sentAt}>{issueDate(i.sentAt)}</time>
                <span>{i.subject}</span>
              </Link>
            </li>
          ))}
        </ol>
      ) : (
        <p className={styles.empty}>The first issue is on its way.</p>
      )}
      <footer className={styles.footer}>
        Published with <a href="https://xem.email">Xem</a> ·{" "}
        <a href={`/n/${a.slug}/feed.xml`}>RSS feed</a>
      </footer>
    </main>
  );
}
