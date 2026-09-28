import { Home } from "@/components/home";
import { getPosts } from "@/lib/blog";
import { siteUrl } from "@/lib/site";
export const metadata = { alternates: { canonical: "/" } };
export default function Page() {
  const posts = getPosts();
  return (
    <>
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{
          __html: JSON.stringify({
            "@context": "https://schema.org",
            "@graph": [
              {
                "@type": "Organization",
                "@id": `${siteUrl}/#organization`,
                name: "Xem",
                url: siteUrl,
                logo: `${siteUrl}/brand/xem-mark.png`,
                sameAs: ["https://github.com/mailxem"],
              },
              {
                "@type": "WebSite",
                "@id": `${siteUrl}/#website`,
                name: "Xem",
                url: siteUrl,
                publisher: { "@id": `${siteUrl}/#organization` },
                inLanguage: "en",
              },
            ],
          }).replace(/</g, "\\u003c"),
        }}
      />
      <Home
        latestPosts={[
          posts.find((p) => p.featured)!,
          posts.find((p) => p.slug === "email-design")!,
          posts.find((p) => p.slug === "email-deliverability")!,
        ]}
      />
    </>
  );
}
