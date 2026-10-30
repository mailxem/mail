import { Home } from "@/components/home";
import { getPosts } from "@/lib/blog";
import { githubUrl, siteUrl } from "@/lib/site";
import { faqs } from "@/lib/content";
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
              {
                "@type": "SoftwareApplication",
                "@id": `${siteUrl}/#software`,
                name: "Xem",
                url: siteUrl,
                description:
                  "Open-source email marketing suite: campaigns, newsletters, automations, forms, CRM, a shared inbox and analytics. Sends through Amazon SES or your own SMTP.",
                applicationCategory: "BusinessApplication",
                applicationSubCategory: "Email marketing",
                operatingSystem: "Web, Linux (self-hosted)",
                isAccessibleForFree: true,
                license: "https://www.gnu.org/licenses/gpl-3.0.html",
                sameAs: [githubUrl],
                offers: {
                  "@type": "Offer",
                  price: "0",
                  priceCurrency: "USD",
                  description: "Free to self-host. Sending costs are paid to your email provider.",
                },
                publisher: { "@id": `${siteUrl}/#organization` },
              },
              {
                "@type": "FAQPage",
                "@id": `${siteUrl}/#faq`,
                mainEntity: faqs.map((f) => ({
                  "@type": "Question",
                  name: f.q,
                  acceptedAnswer: { "@type": "Answer", text: f.a },
                })),
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
