import type { Metadata } from "next";
import Link from "next/link";
import {
  archiveFetch,
  htmlExcerpt,
  issueDate,
  type ArchiveEdition,
} from "@/lib/newsletters/archive";
import { IssueFrame } from "../../issue-frame";
import styles from "../../archive.module.css";

type Params = { params: Promise<{ slug: string; issueId: string }> };

const load = async ({ params }: Params) => {
  const { slug, issueId } = await params;
  return archiveFetch<ArchiveEdition>(
    `${encodeURIComponent(slug)}/issues/${encodeURIComponent(issueId)}`,
  );
};

export async function generateMetadata(props: Params): Promise<Metadata> {
  const e = await load(props);
  const description = htmlExcerpt(e.html);
  return {
    title: `${e.issue.subject} · ${e.newsletter}`,
    description,
    openGraph: {
      title: e.issue.subject,
      description,
      type: "article",
      publishedTime: e.issue.sentAt,
    },
    robots: "index, follow",
  };
}

export default async function NewsletterIssue(props: Params) {
  const e = await load(props);
  return (
    <main className={styles.page}>
      <nav>
        <Link href={`/n/${e.slug}`} className={styles.back}>
          ← {e.newsletter}
        </Link>
      </nav>
      <header className={styles.issueHead}>
        <time dateTime={e.issue.sentAt}>{issueDate(e.issue.sentAt)}</time>
        <h1 className={styles.subject}>{e.issue.subject}</h1>
      </header>
      <IssueFrame html={e.html} title={e.issue.subject} />
      <footer className={styles.footer}>
        Published with <a href="https://xem.email">Xem</a>
      </footer>
    </main>
  );
}
