"use client";
import { workspaceClassName } from "@/lib/workspace-styles";
import { useContext, useState } from "react";
import Link from "next/link";
import { Plus } from "lucide-react";
import { Button } from "@/components/ui/button";
import { TemplatesList } from "@/components/templates/templates-list";
import { StarterGallery } from "@/components/templates/starter-gallery";
import { TemplatesProvider } from "@/app/providers/templates-provider";
import { PreviewTransport, useMarketingQuery } from "@/lib/marketing/api";
import type { Options } from "@/lib/marketing/types";
import { PageHeading, QueryState, FilterTabs } from "./shared";

function PreviewTemplates() {
  const q = useMarketingQuery<Options>("marketing/options");
  return (
    <div className="p-6">
      <QueryState loading={q.isLoading} error={q.error} />
      {q.data?.templates.map((t) => (
        <div className="flex items-center justify-between py-4" key={t.id}>
          <strong>{t.name}</strong>
          <Link href={`/templates/${t.id}/edit`}>Open editor ↗</Link>
        </div>
      ))}
    </div>
  );
}

export function TemplatesPage() {
  const preview = useContext(PreviewTransport);
  const [tab, setTab] = useState("Your templates");
  return (
    <>
      <PageHeading
        title="Templates"
        description="Your existing designs, ready for campaigns, newsletters, and automations."
        action={
          <Button className={workspaceClassName("product-primary")} asChild>
            <Link href="/templates/new">
              <Plus />
              Create Template
            </Link>
          </Button>
        }
      />
      <section className={workspaceClassName("product-panel")}>
        <div className={workspaceClassName("panel-toolbar")}>
          <h2>Template library</h2>
          <FilterTabs
            items={["Your templates", "Predefined templates"]}
            value={tab}
            onChange={setTab}
          />
        </div>
        {tab === "Your templates" ? (
          preview ? (
            <PreviewTemplates />
          ) : (
            <TemplatesProvider>
              <TemplatesList />
            </TemplatesProvider>
          )
        ) : (
          <StarterGallery />
        )}
      </section>
    </>
  );
}
