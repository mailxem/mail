"use client";

import { useMarketingQuery } from "@/lib/marketing/api";
import type {
  FormAnalytics as Analytics,
  LeadForm,
} from "@/lib/marketing/types";
import { Modal, QueryState, Empty } from "./shared";
import { workspaceClassName } from "@/lib/workspace-styles";

export function FormAnalytics({
  form,
  close,
}: {
  form: LeadForm;
  close: () => void;
}) {
  const query = useMarketingQuery<Analytics>(
    `marketing/forms/${form.id}/analytics`,
  );
  const data = query.data;
  return (
    <Modal
      open
      onOpenChange={close}
      title={`${form.Name} · Analytics`}
      description="Understand where visitors start, finish, and choose to save their progress."
      wide
    >
      <QueryState
        loading={query.isLoading}
        error={query.error}
        retry={() => void query.refetch()}
      />
      {data && (
        <div className="space-y-6">
          <dl className="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-5">
            {[
              ["Views", data.views],
              ["Started", data.starts],
              ["Completed", data.completions],
              ["Completion rate", `${(data.completionRate * 100).toFixed(1)}%`],
              ["Saved drafts", data.partials],
            ].map(([label, value]) => (
              <div key={label} className="rounded-lg border p-4">
                <dt className="text-xs text-muted-foreground">{label}</dt>
                <dd className="mt-2 text-2xl font-semibold tabular-nums">
                  {typeof value === "number" ? value.toLocaleString() : value}
                </dd>
              </div>
            ))}
          </dl>
          <p className="text-xs text-muted-foreground">
            Views and starts count browser sessions. Completion rate is
            completed sessions divided by viewed sessions. Saved drafts are
            separate from completed submissions; HTML action submissions may
            have no tracked session.
          </p>
          <section className="space-y-3">
            <h3 className="font-medium">Step-by-step funnel</h3>
            {!data.steps?.length ? (
              <p className="text-sm text-muted-foreground">
                Step activity will appear after visitors use this form.
              </p>
            ) : (
              <div className={workspaceClassName("table-wrap")}>
                <table className={workspaceClassName("product-table")}>
                  <thead>
                    <tr>
                      <th scope="col">Step</th>
                      <th scope="col">Viewed</th>
                      <th scope="col">Completed</th>
                      <th scope="col">Completion</th>
                    </tr>
                  </thead>
                  <tbody>
                    {data.steps.map((step) => (
                      <tr key={step.pageId}>
                        <td>{step.title || step.pageId}</td>
                        <td>{step.views.toLocaleString()}</td>
                        <td>{step.completions.toLocaleString()}</td>
                        <td>
                          {step.views
                            ? `${((step.completions / step.views) * 100).toFixed(1)}%`
                            : "—"}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
            <p className="text-xs text-muted-foreground">
              A step is completed when its visitor finishes the entire form.
              Conditional steps have different audiences; skipped steps do not
              indicate abandonment.
            </p>
          </section>
          <section className="space-y-3">
            <h3 className="font-medium">Sources of completed submissions</h3>
            {!data.sources?.length ? (
              <Empty
                title="No sources yet"
                description="Share links with utm_source to compare your campaigns."
              />
            ) : (
              <div className={workspaceClassName("table-wrap")}>
                <table className={workspaceClassName("product-table")}>
                  <thead>
                    <tr>
                      <th scope="col">Source</th>
                      <th scope="col">Completions</th>
                    </tr>
                  </thead>
                  <tbody>
                    {data.sources.map((source) => (
                      <tr key={source.source}>
                        <td>{source.source || "Direct / unknown"}</td>
                        <td>{source.completions.toLocaleString()}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </section>
        </div>
      )}
    </Modal>
  );
}
