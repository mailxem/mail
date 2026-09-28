"use client";
import { workspaceClassName } from "@/lib/workspace-styles";
import { useState } from "react";
import {
  Plus,
  Eye,
  Copy,
  Pencil,
  MoreVertical,
  FilePenLine,
  CircleCheck,
  Files,
  MousePointerClick,
  Repeat2,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
} from "@/components/ui/dropdown-menu";
import { useMarketing, useMarketingQuery } from "@/lib/marketing/api";
import type { LeadForm, Options } from "@/lib/marketing/types";
import {
  PageHeading,
  Metric,
  Status,
  Empty,
  QueryState,
  Modal,
  FilterTabs,
} from "./shared";
import { toast } from "sonner";
import { FormEditor } from "./form-editor";
import { FormPreview } from "./form-preview";
import { FormShare } from "./form-share";
import { FormAnalytics } from "./form-analytics";
import { FormJourneyPanel } from "./form-journey";
import { getFormDefinition } from "@/lib/marketing/form-definition";
import { formSavePayload } from "@/lib/marketing/form-starters";
export function FormsPage() {
  const query = useMarketingQuery<LeadForm[]>("marketing/forms");
  const options = useMarketingQuery<Options>("marketing/options");
  const { request, refresh } = useMarketing();
  const [filter, setFilter] = useState("All Forms");
  const [editing, setEditing] = useState<LeadForm | null>(null);
  const [open, setOpen] = useState(false);
  const [view, setView] = useState<LeadForm | null>(null);
  const [embed, setEmbed] = useState<LeadForm | null>(null);
  const [submissions, setSubmissions] = useState<LeadForm | null>(null);
  const [analytics, setAnalytics] = useState<LeadForm | null>(null);
  const [journey, setJourney] = useState<LeadForm | null>(null);
  const [pending, setPending] = useState<string | null>(null);
  const forms = query.data || [];
  const active = forms.filter((f) => f.Status === "PUBLISHED").length;
  const total = forms.reduce((n, f) => n + f.SubmissionCount, 0);
  const views = forms.reduce((n, f) => n + f.ViewCount, 0);
  const edit = (form?: LeadForm) => {
    setEditing(form || null);
    setOpen(true);
  };
  const duplicate = async (f: LeadForm) => {
    if (pending) return;
    setPending(f.id);
    try {
      await request("marketing/forms", "POST", {
        ...formSavePayload(f),
        name: `${f.Name.slice(0, 113)} (copy)`,
        status: "DRAFT",
      });
      await refresh();
      toast.success("Form duplicated as a draft");
    } catch (e) {
      toast.error((e as Error).message);
    } finally {
      setPending(null);
    }
  };
  const pause = async (f: LeadForm) => {
    if (pending) return;
    setPending(f.id);
    try {
      await request(`marketing/forms/${f.id}`, "PUT", {
        ...formSavePayload(f),
        status: f.Status === "PUBLISHED" ? "ARCHIVED" : "PUBLISHED",
      });
      await refresh();
    } catch (e) {
      toast.error((e as Error).message);
    } finally {
      setPending(null);
    }
  };
  return (
    <>
      <PageHeading
        title="Forms & Lead Capture"
        description="Turn signups, requests, and feedback into thoughtful email journeys"
        action={
          <Button
            className={workspaceClassName("product-primary")}
            onClick={() => edit()}
          >
            <Plus />
            Create New Form
          </Button>
        }
      />
      <div className={workspaceClassName("metrics-grid")}>
        <Metric
          label="Total Forms"
          value={forms.length.toString().padStart(2, "0")}
          icon={<CircleCheck />}
        />
        <Metric
          label="Total Submissions"
          value={total.toLocaleString()}
          icon={<Files />}
        />
        <Metric
          label="Conversion Rate"
          value={`${views ? ((total / views) * 100).toFixed(1) : "0"}%`}
          icon={<MousePointerClick />}
        />
        <Metric
          label="Active Forms"
          value={active.toString().padStart(2, "0")}
          icon={<Repeat2 />}
          detail={`${forms.length - active} paused / draft`}
        />
      </div>
      <section className={workspaceClassName("product-panel")}>
        <div className={workspaceClassName("panel-toolbar")}>
          <h2>All Forms</h2>
          <FilterTabs
            items={["All Forms", "Active", "Paused"]}
            value={filter}
            onChange={setFilter}
          />
        </div>
        <QueryState
          loading={query.isLoading}
          error={query.error}
          retry={() => query.refetch()}
        />
        {!query.isLoading && !query.error && (
          <div className={workspaceClassName("forms-grid")}>
            {forms
              .filter(
                (f) =>
                  filter === "All Forms" ||
                  (filter === "Active"
                    ? f.Status === "PUBLISHED"
                    : f.Status !== "PUBLISHED"),
              )
              .map((f) => (
                <article className={workspaceClassName("form-card")} key={f.id}>
                  <div className={workspaceClassName("form-card-top")}>
                    <span className={workspaceClassName("form-symbol")}>
                      <FilePenLine />
                    </span>
                    <div className={workspaceClassName("card-actions")}>
                      <button
                        className={workspaceClassName("icon-button")}
                        aria-label={`Preview ${f.Name}`}
                        onClick={() => setView(f)}
                      >
                        <Eye />
                      </button>
                      <button
                        className={workspaceClassName("icon-button")}
                        aria-label={`Duplicate ${f.Name}`}
                        disabled={!!pending}
                        onClick={() => duplicate(f)}
                      >
                        <Copy />
                      </button>
                      <button
                        className={workspaceClassName("icon-button")}
                        aria-label={`Edit ${f.Name}`}
                        onClick={() => edit(f)}
                      >
                        <Pencil />
                      </button>
                      <Status value={f.Status} />
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <button
                            className={workspaceClassName(
                              "icon-button !border-transparent !shadow-none",
                            )}
                            aria-label={`More actions for ${f.Name}`}
                          >
                            <MoreVertical />
                          </button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          <DropdownMenuItem onClick={() => setAnalytics(f)}>
                            View analytics
                          </DropdownMenuItem>
                          <DropdownMenuItem onClick={() => setJourney(f)}>
                            Email journey
                          </DropdownMenuItem>
                          <DropdownMenuItem onClick={() => setSubmissions(f)}>
                            View submissions
                          </DropdownMenuItem>
                          <DropdownMenuItem onClick={() => setEmbed(f)}>
                            Share form
                          </DropdownMenuItem>
                          <DropdownMenuItem
                            disabled={!!pending}
                            onClick={() => pause(f)}
                          >
                            {f.Status === "PUBLISHED"
                              ? "Pause form"
                              : "Publish form"}
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </div>
                  </div>
                  <h3>{f.Name}</h3>
                  <p className={workspaceClassName("form-description")}>
                    {f.description || "A new way to connect with your audience"}
                  </p>
                  <div className={workspaceClassName("form-stats")}>
                    <div>
                      <span>Submissions</span>
                      <strong>{f.SubmissionCount.toLocaleString()}</strong>
                    </div>
                    <div>
                      <span>Conversion Rate</span>
                      <strong>
                        {f.ViewCount
                          ? ((f.SubmissionCount / f.ViewCount) * 100).toFixed(1)
                          : "0"}
                        %
                      </strong>
                    </div>
                  </div>
                </article>
              ))}
          </div>
        )}
        {!query.isLoading && !query.error && forms.length === 0 && (
          <Empty
            title="A warm welcome starts here"
            description="Build your first form to welcome subscribers, learn about customers, or collect feedback."
            action={
              <Button
                className={workspaceClassName("product-primary")}
                onClick={() => edit()}
              >
                <Plus />
                Create your first form
              </Button>
            }
          />
        )}
      </section>
      <FormEditor
        key={`${editing?.id || "new"}-${open}`}
        open={open}
        onOpenChange={setOpen}
        form={editing}
        options={options.data}
        forms={forms}
      />
      <Modal
        open={!!view}
        onOpenChange={() => setView(null)}
        title="Form preview"
        description="This is how your form will appear to visitors."
      >
        {view && (
          <FormPreview
            name={view.Name}
            description={view.description}
            definition={getFormDefinition(view)}
            button={view.SubmitButtonText}
            theme={view.theme}
          />
        )}
      </Modal>
      {embed && <FormShare form={embed} close={() => setEmbed(null)} />}
      {analytics && (
        <FormAnalytics form={analytics} close={() => setAnalytics(null)} />
      )}
      {journey && (
        <Modal
          open
          onOpenChange={() => setJourney(null)}
          title={`${journey.Name} · Email journey`}
          description="Review the emails and routing that follow this form before activating your journey."
          wide
        >
          <FormJourneyPanel formId={journey.id} formName={journey.Name} />
        </Modal>
      )}
      {submissions && (
        <Submissions form={submissions} close={() => setSubmissions(null)} />
      )}
    </>
  );
}
function Submissions({ form, close }: { form: LeadForm; close: () => void }) {
  const q = useMarketingQuery<
    { id: string; EmailAddress: string; SubmittedAt: string }[]
  >(`marketing/forms/${form.id}/submissions`);
  return (
    <Modal
      open
      onOpenChange={close}
      title={`${form.Name} submissions`}
      description="The 200 most recent submissions, including repeat signups."
    >
      <QueryState
        loading={q.isLoading}
        error={q.error}
        retry={() => void q.refetch()}
      />
      {q.data?.length === 0 && (
        <Empty
          title="No submissions yet"
          description="Share your form to start the conversation."
        />
      )}
      <div className={workspaceClassName("table-wrap")}>
        <table className={workspaceClassName("product-table")}>
          <thead>
            <tr>
              <th>Email</th>
              <th>Submitted</th>
            </tr>
          </thead>
          <tbody>
            {q.data?.map((s) => (
              <tr key={s.id}>
                <td>{s.EmailAddress}</td>
                <td>{new Date(s.SubmittedAt).toLocaleString()}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </Modal>
  );
}
