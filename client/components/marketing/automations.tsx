"use client";
import { workspaceClassName } from "@/lib/workspace-styles";
import { Fragment, useCallback, useEffect, useMemo, useState } from "react";
import {
  ReactFlow,
  Background,
  Controls,
  Handle,
  Position,
  useNodesState,
  useEdgesState,
  addEdge,
  MarkerType,
  type Node,
  type Edge,
  type Connection,
} from "@xyflow/react";
import "@xyflow/react/dist/style.css";
import {
  Plus,
  Workflow as WorkflowIcon,
  Mail,
  Clock,
  GitBranch,
  Flag,
  Zap,
  ArrowLeft,
  Check,
  Play,
  Pause,
  Save,
  Trash2,
  MousePointer2,
  MoreHorizontal,
  ArrowUpRight,
  History,
  ShieldCheck,
  ChevronRight,
  Tags,
  ListPlus,
  ContactRound,
  Shuffle,
  Variable,
  Hourglass,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { GmailClipNotice } from "@/components/campaigns/gmail-clip-notice";
import { useMarketing, useMarketingQuery } from "@/lib/marketing/api";
import type {
  Workflow,
  Options,
  WorkflowNode,
  Contact,
  LeadForm,
  StepStats,
  WorkflowVersion,
} from "@/lib/marketing/types";
import {
  PageHeading,
  Metric,
  QueryState,
  Empty,
  Modal,
  Field,
  Status,
  FilterTabs,
} from "./shared";
import { toast } from "sonner";
import { AutomationWindowSetting } from "@/components/settings/sending-rules-settings";
import {
  ActionSettings,
  ConditionSettings,
  DateTriggerSettings,
  WaitSettings,
  WaitUntilSettings,
  waitEvents,
} from "./automation-settings";
const icons: Record<string, typeof Zap> = {
  START: Zap,
  EMAIL: Mail,
  WAIT: Clock,
  WAIT_UNTIL: Hourglass,
  CONDITION: GitBranch,
  EXIT: Flag,
  TAG: Tags,
  ADD_TO_LIST: ListPlus,
  UPDATE_SUBSCRIBER: ContactRound,
  PERCENTAGE_SPLIT: Shuffle,
  SET_VARIABLE: Variable,
};
const labels: Record<string, string> = {
  START: "Trigger",
  EMAIL: "Send an email",
  WAIT: "Wait for a while",
  WAIT_UNTIL: "Wait until something happens",
  CONDITION: "Split into paths",
  EXIT: "Finish workflow",
  TAG: "Manage tags",
  ADD_TO_LIST: "Move to a list",
  UPDATE_SUBSCRIBER: "Update contact",
  PERCENTAGE_SPLIT: "Percentage split",
  SET_VARIABLE: "Set a variable",
};
// Steps with two labelled exits, and the labels of those exits.
const branchPaths: Record<string, [string, string]> = {
  CONDITION: ["true", "false"],
  PERCENTAGE_SPLIT: ["A", "B"],
  WAIT_UNTIL: ["happened", "timeout"],
};
const triggerLabels: Record<string, string> = {
  manual: "Started manually",
  "contact.created": "New contact added",
  "email.opened": "Email opened",
  "email.clicked": "Email link clicked",
  "form.completed": "Form completed",
  "contact.tag_added": "Tag added to a contact",
  "list.subscriber_added": "Contact joins a list",
  "segment.entered": "Contact enters a segment",
  "date.reached": "A contact date comes round",
  "event.received": "Event from your app",
};
const stepDescriptions: Record<string, string> = {
  EMAIL: "Template, sender and subject",
  WAIT: "Duration or a specific date",
  WAIT_UNTIL: "An open, click, tag or event, or a time limit",
  CONDITION: "Match all or any rules",
  TAG: "Add or remove contact tags",
  ADD_TO_LIST: "Change the contact’s list",
  UPDATE_SUBSCRIBER: "Update profile fields",
  PERCENTAGE_SPLIT: "Send contacts down A/B paths",
  SET_VARIABLE: "Remember a value for later steps",
};
function StepNode({
  data,
  selected,
}: {
  data: Record<string, any>;
  selected?: boolean;
}) {
  const Icon = icons[data.kind] || Zap;
  return (
    <div
      className={workspaceClassName(
        `workflow-step ${selected ? "selected" : ""}`,
      )}
    >
      {data.kind !== "START" && (
        <Handle type="target" position={Position.Top} />
      )}
      <div
        className={workspaceClassName(
          `step-icon step-${data.kind.toLowerCase()}`,
        )}
      >
        <Icon size={19} />
      </div>
      <div>
        <small>
          {data.kind === "START"
            ? "WHEN THIS HAPPENS"
            : data.kind === "EXIT"
              ? "ALL DONE"
              : `THEN DO THIS`}
        </small>
        <strong>{data.title || data.label || labels[data.kind]}</strong>
        <p>{data.summary || "Click to configure this step"}</p>
        {data.stats && (
          <p className="mt-1 text-[11px] text-muted-foreground">
            {count(data.stats.entered)} entered · {count(data.stats.waiting)}{" "}
            waiting
            {data.kind === "EMAIL" &&
              ` · ${percent(data.stats.openRate)} opened · ${percent(data.stats.clickRate)} clicked`}
          </p>
        )}
      </div>
      <ChevronRight size={15} className={workspaceClassName("step-chevron")} />
      {branchPaths[data.kind]
        ? branchPaths[data.kind].map((path, index) => (
            <Handle
              key={path}
              id={path}
              type="source"
              position={Position.Bottom}
              style={{ left: index ? "75%" : "25%" }}
              title={`Path ${path}`}
            />
          ))
        : data.kind !== "EXIT" && (
            <Handle type="source" position={Position.Bottom} />
          )}
    </div>
  );
}
const nodeTypes = { step: StepNode };
const count = (n: number) => n.toLocaleString();
const percent = (rate: number) => `${Math.round(rate * 1000) / 10}%`;
function stepResults(s: StepStats | undefined, kind: string) {
  const z = s || ({} as Partial<StepStats>);
  const rows: [string, string][] = [
    ["Entered", count(z.entered || 0)],
    ["Waiting here", count(z.waiting || 0)],
    ["Moved on", count(z.movedOn || 0)],
    ["Exited here", count(z.exited || 0)],
    ["Reached the goal", count(z.conversions || 0)],
  ];
  if (kind === "EMAIL")
    rows.push(
      ["Sent", count(z.sent || 0)],
      [
        "Delivered",
        z.delivered == null ? "Not reported" : count(z.delivered),
      ],
      ["Unique opens", `${count(z.opens || 0)} · ${percent(z.openRate || 0)}`],
      ["Unique clicks", `${count(z.clicks || 0)} · ${percent(z.clickRate || 0)}`],
      [
        "Unsubscribes",
        `${count(z.unsubscribes || 0)} · ${percent(z.unsubscribeRate || 0)}`,
      ],
    );
  return rows;
}
const toEdge = (e: Workflow["edges"][number], i: number): Edge => ({
  id: e.id || `edge-${i}`,
  source: e.sourceId,
  target: e.targetId,
  label: e.label,
  type: "smoothstep",
  markerEnd: { type: MarkerType.ArrowClosed, width: 12, height: 12 },
  style: { stroke: "#bcb0d0", strokeWidth: 1.5 },
});
function starter(kind = "blank"): Workflow {
  const start = crypto.randomUUID(),
    exit = crypto.randomUUID(),
    email = crypto.randomUUID(),
    wait = crypto.randomUUID();
  const nodes: WorkflowNode[] = [
    { id: start, type: "START", data: { position: { x: 0, y: 0 } } },
  ];
  if (kind !== "blank")
    nodes.push({
      id: email,
      type: "EMAIL",
      data: { position: { x: 0, y: 165 } },
    });
  if (kind === "nurture")
    nodes.push({
      id: wait,
      type: "WAIT",
      data: { duration: "48h", position: { x: 0, y: 330 } },
    });
  nodes.push({
    id: exit,
    type: "EXIT",
    data: { position: { x: 0, y: 165 * nodes.length } },
  });
  return {
    id: "",
    name:
      kind === "welcome"
        ? "A warm welcome"
        : kind === "nurture"
          ? "Stay in the conversation"
          : "Untitled automation",
    description: "",
    triggerEvent: kind === "blank" ? "manual" : "contact.created",
    isActive: false,
    nodes,
    edges: nodes
      .slice(1)
      .map((n, i) => ({ sourceId: nodes[i].id, targetId: n.id, label: "" })),
    updatedAt: "",
  };
}
export function AutomationsPage() {
  const q = useMarketingQuery<Workflow[] | { data: Workflow[] }>("automations");
  const [editing, setEditing] = useState<Workflow | null>(null);
  const [tab, setTab] = useState("All automations");
  const rows = Array.isArray(q.data) ? q.data : q.data?.data || [];
  if (editing)
    return (
      <AutomationBuilder
        key={editing.id || editing.name}
        initial={editing}
        close={() => setEditing(null)}
      />
    );
  return (
    <>
      <PageHeading
        title="A thoughtful journey. On autopilot."
        description="Turn the right moments into meaningful conversations."
        action={
          <Button
            className={workspaceClassName("product-primary")}
            onClick={() => setEditing(starter())}
          >
            <Plus />
            Create Automation
          </Button>
        }
      />
      <div className={workspaceClassName("automation-starters")}>
        <button onClick={() => setEditing(starter("welcome"))}>
          <span className={workspaceClassName("starter-icon")}>
            <Mail />
          </span>
          <div>
            <strong>Welcome new subscribers</strong>
            <p>A warm introduction, right on time.</p>
          </div>
          <ArrowUpRight />
        </button>
        <button onClick={() => setEditing(starter("nurture"))}>
          <span className={workspaceClassName("starter-icon peach")}>
            <Clock />
          </span>
          <div>
            <strong>Build a lasting connection</strong>
            <p>Bring a little intention to your follow-up.</p>
          </div>
          <ArrowUpRight />
        </button>
      </div>
      <section className={workspaceClassName("product-panel")}>
        <div className={workspaceClassName("panel-toolbar")}>
          <h2>
            Your automations{" "}
            <small className="text-muted-foreground text-xs ml-2">
              {rows.length}
            </small>
          </h2>
          <FilterTabs
            items={["All automations", "Active", "Drafts"]}
            value={tab}
            onChange={setTab}
          />
        </div>
        <QueryState
          loading={q.isLoading}
          error={q.error}
          retry={() => q.refetch()}
        />
        {!q.isLoading && !q.error && rows.length === 0 && (
          <Empty
            title="Make every moment count"
            description="Start with a ready-made journey or build a workflow from scratch."
            action={
              <Button
                className={workspaceClassName("product-primary")}
                onClick={() => setEditing(starter())}
              >
                <Plus />
                Build your first automation
              </Button>
            }
          />
        )}
        {rows.length > 0 && (
        <div className={workspaceClassName("table-wrap")}>
          <table className={workspaceClassName("product-table")}>
            <thead>
              <tr>
                <th>Automation</th>
                <th>Status</th>
                <th>Starts when</th>
                <th>Steps</th>
                <th>Last edited</th>
                <th />
              </tr>
            </thead>
            <tbody>
              {rows
                .filter(
                  (a) =>
                    tab === "All automations" ||
                    (tab === "Active" ? a.isActive : !a.isActive),
                )
                .map((a) => (
                  <tr key={a.id}>
                    <td>
                      <button
                        className={workspaceClassName(
                          "newsletter-title text-left",
                        )}
                        onClick={() => setEditing(a)}
                      >
                        <span>
                          <WorkflowIcon size={18} />
                        </span>
                        <div>
                          <strong>{a.name}</strong>
                          <small>
                            {a.description || "A connected customer journey"}
                          </small>
                        </div>
                      </button>
                    </td>
                    <td>
                      <Status value={a.isActive ? "PUBLISHED" : "DRAFT"} />
                    </td>
                    <td>
                      {triggerLabels[a.triggerEvent] || "Started manually"}
                    </td>
                    <td>{a.nodes?.length || 0} steps</td>
                    <td>
                      {a.updatedAt
                        ? new Date(a.updatedAt).toLocaleDateString()
                        : "—"}
                    </td>
                    <td>
                      <button
                        aria-label={`Edit ${a.name}`}
                        className={workspaceClassName("icon-button")}
                        onClick={() => setEditing(a)}
                      >
                        <ArrowUpRight />
                      </button>
                    </td>
                  </tr>
                ))}
            </tbody>
          </table>
        </div>
        )}
      </section>
    </>
  );
}
export function AutomationBuilder({
  initial,
  close,
}: {
  initial: Workflow;
  close: () => void;
}) {
  const options = useMarketingQuery<Options>("marketing/options");
  const forms = useMarketingQuery<LeadForm[]>("marketing/forms");
  const [trigger, setTrigger] = useState(initial.triggerEvent || "manual");
  const tags = useMarketingQuery<{ id: string; name: string }[]>(
    "marketing/tags",
    trigger === "contact.tag_added",
  );
  const segments = useMarketingQuery<{
    data: { id: string; name: string; isDynamic: boolean }[];
  }>("segments?limit=100", trigger === "segment.entered");
  const { request, refresh } = useMarketing();
  const [id, setID] = useState(initial.id);
  const [name, setName] = useState(initial.name);
  const [description, setDescription] = useState(initial.description || "");
  const [stepSearch, setStepSearch] = useState("");
  const [sendWindow, setSendWindow] = useState(initial.sendWindow ?? null);
  const [active, setActive] = useState(initial.isActive);
  // The canvas is always the draft; contacts run on the live version.
  const [liveVersion, setLiveVersion] = useState(initial.liveVersion || 0);
  const [hasDraft, setHasDraft] = useState(!!initial.hasDraft);
  const [publishing, setPublishing] = useState(false);
  const [moveRunning, setMoveRunning] = useState(false);
  const [versions, setVersions] = useState(false);
  const [range, setRange] = useState(30);
  const from = useMemo(
    () => new Date(Date.now() - range * 864e5).toISOString(),
    [range],
  );
  const stats = useMarketingQuery<{ steps: StepStats[] }>(
    `automations/${id}/stats?from=${encodeURIComponent(from)}`,
    !!id && liveVersion > 0,
  );
  const statsByNode = new Map(
    (stats.data?.steps || []).map((step) => [step.nodeId, step]),
  );
  // Older workflows without a rule behave as "again after it finishes".
  const [reentryMode, setReentryMode] = useState(
    initial.reentryMode || (initial.id ? "after_finish" : "once"),
  );
  const [reentryDays, setReentryDays] = useState(initial.reentryDays || 30);
  const [exitOnGoal, setExitOnGoal] = useState(initial.exitOnGoal !== false);
  const [selected, setSelected] = useState(
    initial.nodes?.find((n) => n.type === "START")?.id || "",
  );
  const [busy, setBusy] = useState(false);
  const [dirty, setDirty] = useState(false);
  const [error, setError] = useState("");
  const [runs, setRuns] = useState(false);
  const [test, setTest] = useState(false);
  const [leave, setLeave] = useState(false);
  const toFlow = (n: WorkflowNode, i: number): Node => ({
    id: n.id,
    type: "step",
    position: n.data?.position || { x: 0, y: i * 165 },
    data: { ...n.data, kind: n.type, label: labels[n.type] },
  });
  const [nodes, setNodes, onNodesChange] = useNodesState(
    (initial.nodes || []).filter((n) => !(n as any).isDeleted).map(toFlow),
  );
  const [edges, setEdges, onEdgesChange] = useEdgesState(
    (initial.edges || []).map(toEdge),
  );
  const load = (w: Workflow) => {
    setNodes((w.nodes || []).filter((n) => !(n as any).isDeleted).map(toFlow));
    setEdges((w.edges || []).map(toEdge));
    setName(w.name);
    setDescription(w.description || "");
    setTrigger(w.triggerEvent || "manual");
    setHasDraft(!!w.hasDraft);
    setDirty(false);
  };
  const current = nodes.find((n) => n.id === selected);
  const config = current?.data || {};
  const kind = String(config.kind || "");
  useEffect(() => {
    const handler = (e: BeforeUnloadEvent) => {
      if (dirty) {
        e.preventDefault();
        e.returnValue = "";
      }
    };
    window.addEventListener("beforeunload", handler);
    return () => window.removeEventListener("beforeunload", handler);
  }, [dirty]);
  const update = (values: Record<string, any>) => {
    setNodes((ns) =>
      ns.map((n) =>
        n.id === selected ? { ...n, data: { ...n.data, ...values } } : n,
      ),
    );
    setDirty(true);
  };
  const serialized = () => ({
    name,
    description,
    triggerEvent: trigger,
    reentryMode,
    reentryDays: reentryMode === "after_days" ? reentryDays : 0,
    exitOnGoal,
    sendWindow,
    nodes: nodes.map((n) => {
      const { kind, label, summary, ...data } = n.data;
      return { id: n.id, type: kind, data: { ...data, position: n.position } };
    }),
    edges: edges.map((e) => ({
      sourceId: e.source,
      targetId: e.target,
      label: typeof e.label === "string" ? e.label : "",
    })),
  });
  async function save() {
    const result = await request<Workflow>(
      `automations${id ? `/${id}` : ""}`,
      id ? "PUT" : "POST",
      serialized(),
    );
    setID(result.id);
    setDirty(false);
    await refresh();
    return result.id;
  }
  async function action(
    type: "save" | "validate" | "publish" | "pause" | "resume" | "discard",
  ) {
    setBusy(true);
    setError("");
    try {
      if (type === "pause") {
        await request(`automations/${id}/deactivate`, "POST", {});
        setActive(false);
        await refresh();
        toast.success("Workflow paused");
      } else if (type === "resume") {
        await request(`automations/${id}/activate`, "POST", {});
        setActive(true);
        await refresh();
        toast.success(`Version ${liveVersion} is running again`);
      } else if (type === "discard") {
        load(await request<Workflow>(`automations/${id}/discard`, "POST", {}));
        toast.success(`Draft discarded; showing version ${liveVersion}`);
      } else if (type === "validate") {
        await request("automations/validate", "POST", serialized());
        toast.success("All steps and connections are valid");
      } else {
        const saved = await save();
        if (type === "publish") {
          const version = await request<{ number: number }>(
            `automations/${saved}/publish`,
            "POST",
            { moveRunning },
          );
          setLiveVersion(version.number);
          setHasDraft(false);
          setMoveRunning(false);
          if (!active) {
            await request(`automations/${saved}/activate`, "POST", {});
            setActive(true);
          }
          await refresh();
          toast.success(`Version ${version.number} is live`);
        } else {
          setHasDraft(liveVersion > 0);
          toast.success("Draft saved");
        }
      }
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setBusy(false);
    }
  }
  async function restore(number: number) {
    setVersions(false);
    try {
      load(
        await request<Workflow>(
          `automations/${id}/versions/${number}/restore`,
          "POST",
          {},
        ),
      );
      toast.success(`Version ${number} restored as the draft`);
    } catch (e) {
      setError((e as Error).message);
    }
  }
  const addStep = (type: string) => {
    const incoming =
      current?.data.kind === "EXIT"
        ? edges.filter((e) => e.target === current.id)
        : [];
    if (current?.data.kind === "EXIT" && incoming.length !== 1) {
      setError("Select a step on one of the incoming paths first.");
      return;
    }
    const source =
      current?.data.kind === "EXIT"
        ? nodes.find((n) => n.id === incoming[0]?.source)
        : current;
    if (!source) return;
    const nextEdges =
      current?.data.kind === "EXIT"
        ? incoming
        : edges.filter((e) => e.source === source.id);
    if (nextEdges.length !== 1) {
      setError("Select a step within a branch before adding another step.");
      return;
    }
    setError("");
    const nodeID = crypto.randomUUID();
    const next = nextEdges[0];
    const newNode: Node = {
      id: nodeID,
      type: "step",
      position: { x: source.position.x, y: source.position.y + 165 },
      data: {
        kind: type,
        label: labels[type],
        ...(type === "WAIT" ? { duration: "24h" } : {}),
        ...(type === "WAIT_UNTIL" ? { event: "opened", timeout: "72h" } : {}),
        ...(type === "TAG" ? { action: "add", tags: [] } : {}),
        ...(type === "PERCENTAGE_SPLIT" ? { percentage: 50 } : {}),
        ...(type === "SET_VARIABLE"
          ? { variable: "workflow_stage", value: "" }
          : {}),
        ...(type === "UPDATE_SUBSCRIBER" ? { fields: { company: "" } } : {}),
        ...(type === "CONDITION"
          ? {
              operator: "AND",
              conditions: [
                { variable: "contact_email", operator: "contains", value: "" },
              ],
            }
          : {}),
      },
    };
    let newNodes = [
      ...nodes.map((n) =>
        n.position.y > source.position.y
          ? { ...n, position: { ...n.position, y: n.position.y + 165 } }
          : n,
      ),
      newNode,
    ];
    let newEdges = edges.filter((e) => e.id !== next?.id);
    newEdges.push({
      id: crypto.randomUUID(),
      source: source.id,
      target: nodeID,
      label: next?.label,
      type: "smoothstep",
    });
    if (branchPaths[type]) {
      const falseID = crypto.randomUUID();
      newNodes.push({
        id: falseID,
        type: "step",
        position: { x: source.position.x + 380, y: source.position.y + 340 },
        data: { kind: "EXIT", label: labels.EXIT },
      });
      newEdges.push({
        id: crypto.randomUUID(),
        source: nodeID,
        target: falseID,
        label: branchPaths[type][1],
        type: "smoothstep",
      });
      if (next)
        newEdges.push({
          ...next,
          id: crypto.randomUUID(),
          source: nodeID,
          label: branchPaths[type][0],
        });
    } else if (next)
      newEdges.push({
        ...next,
        id: crypto.randomUUID(),
        source: nodeID,
        label: "",
      });
    setNodes(newNodes);
    setEdges(newEdges);
    setSelected(nodeID);
    setDirty(true);
  };
  const remove = () => {
    if (!current || kind === "START" || kind === "EXIT") return;
    const ins = edges.filter((e) => e.target === selected);
    const outs = edges.filter((e) => e.source === selected);
    if (ins.length !== 1 || outs.length !== 1) {
      setError(
        "Reconnect or remove this branch’s connections before deleting the step.",
      );
      return;
    }
    setEdges([
      ...edges.filter((e) => e.source !== selected && e.target !== selected),
      { ...ins[0], id: crypto.randomUUID(), target: outs[0].target },
    ]);
    setNodes(nodes.filter((n) => n.id !== selected));
    setSelected(ins[0].source);
    setDirty(true);
  };
  const displayNodes = nodes.map((n) => ({
    ...n,
    selected: n.id === selected,
    data: {
      ...n.data,
      stats: statsByNode.get(n.id),
      summary:
        n.data.kind === "START"
          ? triggerLabels[trigger]
          : n.data.kind === "EMAIL"
            ? options.data?.templates.find((t) => t.id === n.data.templateId)
                ?.name || "Choose a template and sender"
            : n.data.kind === "WAIT"
              ? n.data.until
                ? `Wait until ${new Date(String(n.data.until)).toLocaleString()}`
                : `Pause for ${n.data.duration || "…"}`
              : n.data.kind === "WAIT_UNTIL"
                ? `${waitEvents[String(n.data.event)] || "Choose an event"} · at most ${n.data.timeout || "…"}`
                : n.data.kind === "CONDITION"
                  ? `${(n.data.conditions as any[])?.length || 0} rules · ${n.data.operator === "OR" ? "match any" : "match all"}`
                  : n.data.kind === "PERCENTAGE_SPLIT"
                    ? `A: ${n.data.percentage}% · B: ${100 - Number(n.data.percentage)}%`
                    : n.data.kind === "TAG"
                      ? `${n.data.action === "remove" ? "Remove" : "Add"}: ${((n.data.tags as string[]) || []).join(", ") || "choose tags"}`
                      : n.data.kind === "ADD_TO_LIST"
                        ? options.data?.lists.find(
                            (l) => l.id === n.data.listId,
                          )?.name || "Choose a destination list"
                        : n.data.kind === "UPDATE_SUBSCRIBER"
                          ? `Update ${Object.keys(n.data.fields || {}).length} profile fields`
                          : n.data.kind === "SET_VARIABLE"
                            ? `${n.data.variable || "Choose a variable"} = ${n.data.value || "…"}`
                            : "End this journey",
    },
  }));
  return (
    <div className={workspaceClassName("automation-editor")}>
      <div className={workspaceClassName("builder-toolbar")}>
        <button
          className={workspaceClassName("icon-button")}
          aria-label="Back to automations"
          onClick={() => (dirty ? setLeave(true) : close())}
        >
          <ArrowLeft />
        </button>
        <input
          aria-label="Automation name"
          value={name}
          maxLength={120}
          onChange={(e) => {
            setName(e.target.value);
            setDirty(true);
          }}
        />
        <Status
          value={active ? "PUBLISHED" : liveVersion ? "ARCHIVED" : "DRAFT"}
        />
        <span className={workspaceClassName("save-state")}>
          {liveVersion > 0 && `Live: version ${liveVersion} · `}
          {dirty
            ? "Unsaved changes"
            : liveVersion > 0
              ? hasDraft
                ? "Draft not published"
                : "Draft matches live"
              : id
                ? "All changes saved"
                : "New workflow"}
        </span>
        <div className={workspaceClassName("builder-toolbar-actions")}>
          <Button
            variant="outline"
            size="sm"
            disabled={!id}
            onClick={() => setRuns(true)}
          >
            <History />
            History
          </Button>
          <Button
            variant="outline"
            size="sm"
            disabled={!id || !liveVersion}
            onClick={() => setVersions(true)}
          >
            Versions
          </Button>
          <Button
            variant="outline"
            size="sm"
            disabled={busy}
            onClick={() => action("validate")}
          >
            <ShieldCheck />
            Validate
          </Button>
          {active ? (
            <>
              <Button variant="outline" size="sm" onClick={() => setTest(true)}>
                <Play />
                Run
              </Button>
              <Button
                variant="outline"
                disabled={busy}
                onClick={() => action("pause")}
              >
                <Pause />
                Pause
              </Button>
            </>
          ) : (
            liveVersion > 0 &&
            !hasDraft &&
            !dirty && (
              <Button
                variant="outline"
                disabled={busy}
                onClick={() => action("resume")}
              >
                <Play />
                Resume
              </Button>
            )
          )}
          {liveVersion > 0 && hasDraft && !dirty && (
            <Button
              variant="outline"
              disabled={busy}
              onClick={() => action("discard")}
            >
              Discard draft
            </Button>
          )}
          <Button
            variant="outline"
            disabled={busy}
            onClick={() => action("save")}
          >
            <Save />
            Save draft
          </Button>
          <Button
            className={workspaceClassName("product-primary")}
            disabled={busy}
            onClick={() =>
              liveVersion > 0 ? setPublishing(true) : action("publish")
            }
          >
            <Play />
            Publish
          </Button>
        </div>
      </div>
      {error && (
        <div
          role="alert"
          className={workspaceClassName("product-error builder-error")}
        >
          {error}
        </div>
      )}
      {liveVersion > 0 && (hasDraft || dirty) && (
        <div className={workspaceClassName("builder-notice")}>
          You’re editing a draft. Contacts keep running on version{" "}
          {liveVersion} until you publish.
        </div>
      )}
      <div className={workspaceClassName("builder-workspace")}>
        <aside className={workspaceClassName("step-palette")}>
          <div className={workspaceClassName("eyebrow")}>
            Build your journey
          </div>
          <h3>One step at a time.</h3>
          <p>Add an action after the selected step.</p>
          <input
            aria-label="Search steps"
            placeholder="Search steps…"
            value={stepSearch}
            onChange={(e) => setStepSearch(e.target.value)}
            className="w-full rounded-md border bg-background px-2 py-2 text-xs mb-3"
          />
          {Object.keys(stepDescriptions)
            .filter((type) =>
              `${labels[type]} ${stepDescriptions[type]}`
                .toLowerCase()
                .includes(stepSearch.toLowerCase()),
            )
            .map((type) => {
              const Icon = icons[type];
              return (
                <button
                  key={type}
                  title={labels[type]}
                  aria-label={`Add ${labels[type]}`}
                  onClick={() => addStep(type)}
                >
                  <span
                    className={workspaceClassName(
                      `step-icon step-${type.toLowerCase()}`,
                    )}
                  >
                    <Icon size={17} />
                  </span>
                  <span>
                    <strong>{labels[type]}</strong>
                    <small>{stepDescriptions[type]}</small>
                  </span>
                  <Plus size={14} />
                </button>
              );
            })}
          <div className={workspaceClassName("palette-hint")}>
            <MousePointer2 size={17} />
            <p>
              Drag steps to arrange them. Connect the dots to change the
              journey.
            </p>
          </div>
        </aside>
        <div className={workspaceClassName("workflow-canvas")}>
          <div className={workspaceClassName("canvas-label")}>
            <WorkflowIcon size={14} />
            WORKFLOW EDITOR
          </div>
          <ReactFlow
            nodes={displayNodes}
            edges={edges.map((e) => ({
              ...e,
              sourceHandle: branchPaths[
                String(nodes.find((n) => n.id === e.source)?.data.kind)
              ]
                ? String(e.label || "")
                : undefined,
            }))}
            nodeTypes={nodeTypes}
            onNodesChange={(changes) => {
              const selection = changes.find(
                (change) => change.type === "select" && change.selected,
              );
              if (selection && "id" in selection) setSelected(selection.id);
              onNodesChange(changes);
              if (changes.some((c) => c.type === "position" && c.dragging))
                setDirty(true);
            }}
            onEdgesChange={(changes) => {
              onEdgesChange(changes);
              if (changes.some((c) => c.type === "remove")) setDirty(true);
            }}
            onConnect={(connection: Connection) => {
              if (connection.source === connection.target) return;
              setEdges((es) =>
                addEdge(
                  {
                    ...connection,
                    id: crypto.randomUUID(),
                    type: "smoothstep",
                    label: connection.sourceHandle || "",
                  },
                  es,
                ),
              );
              setDirty(true);
            }}
            onNodeClick={(_, n) => setSelected(n.id)}
            fitView
            fitViewOptions={{ padding: 0.4, maxZoom: 0.9 }}
            minZoom={0.25}
            maxZoom={1.5}
            deleteKeyCode={null}
          >
            <Background gap={20} size={1.2} color="#e5deed" />
            <Controls showInteractive={false} />
          </ReactFlow>
        </div>
        <aside className={workspaceClassName("step-inspector")}>
          <div className={workspaceClassName("inspector-title")}>
            <span
              className={workspaceClassName(
                `step-icon step-${kind.toLowerCase()}`,
              )}
            >
              {kind &&
                (() => {
                  const Icon = icons[kind] || Zap;
                  return <Icon size={20} />;
                })()}
            </span>
            <div>
              <small>STEP SETTINGS</small>
              <h3>{labels[kind] || "Select a step"}</h3>
            </div>
          </div>
          {liveVersion > 0 && current && (
            <section
              aria-label="Step results"
              className="mb-4 rounded-md border p-3 text-xs"
            >
              <div className="mb-2 flex items-center justify-between gap-2">
                <strong>Results</strong>
                <select
                  aria-label="Results period"
                  value={range}
                  onChange={(e) => setRange(Number(e.target.value))}
                  className="rounded border bg-background px-1 py-0.5"
                >
                  {[7, 30, 90].map((days) => (
                    <option key={days} value={days}>
                      Last {days} days
                    </option>
                  ))}
                </select>
              </div>
              <dl className="grid grid-cols-2 gap-x-3 gap-y-1">
                {stepResults(statsByNode.get(current.id), kind).map(
                  ([label, value]) => (
                    <Fragment key={label}>
                      <dt className="text-muted-foreground">{label}</dt>
                      <dd className="text-right font-medium">{value}</dd>
                    </Fragment>
                  ),
                )}
              </dl>
              <p className="mt-2 text-muted-foreground">
                Contacts who entered the workflow in this period. Rates are
                over sent emails.
              </p>
            </section>
          )}
          <fieldset className={workspaceClassName("product-form")}>
            {current && (
              <Field label="Step name" hint="Optional label for your team.">
                <input
                  maxLength={100}
                  value={String(config.title || "")}
                  placeholder={labels[kind]}
                  onChange={(e) => update({ title: e.target.value })}
                />
              </Field>
            )}
            {kind === "START" && (
              <>
                <Field label="Workflow description">
                  <textarea
                    maxLength={1000}
                    value={description}
                    onChange={(e) => {
                      setDescription(e.target.value);
                      setDirty(true);
                    }}
                  />
                </Field>
                <Field label="Start this workflow when">
                  <select
                    value={trigger}
                    onChange={(e) => {
                      setTrigger(e.target.value);
                      setDirty(true);
                    }}
                  >
                    {Object.entries(triggerLabels).map(([value, label]) => (
                      <option key={value} value={value}>
                        {label}
                      </option>
                    ))}
                  </select>
                </Field>
                {trigger === "form.completed" && (
                  <Field
                    label="Which form?"
                    hint="Only opted-in subscribers completing this form enter the journey."
                  >
                    <select
                      value={String(config.formId || "")}
                      onChange={(e) => update({ formId: e.target.value })}
                    >
                      <option value="">Choose a form</option>
                      {forms.data?.map((form) => (
                        <option key={form.id} value={form.id}>
                          {form.Name}
                        </option>
                      ))}
                    </select>
                  </Field>
                )}
                {trigger === "contact.tag_added" && (
                  <Field
                    label="Which tag?"
                    hint="Starts when this tag is newly added, by hand, by a workflow step or with a new contact."
                  >
                    <select
                      value={String(config.tagId || "")}
                      onChange={(e) => update({ tagId: e.target.value })}
                    >
                      <option value="">Choose a tag</option>
                      {tags.data?.map((tag) => (
                        <option key={tag.id} value={tag.id}>
                          {tag.name}
                        </option>
                      ))}
                    </select>
                  </Field>
                )}
                {trigger === "list.subscriber_added" && (
                  <Field
                    label="Which list?"
                    hint="Starts for new contacts in this list, imports included, and contacts moved into it."
                  >
                    <select
                      value={String(config.listId || "")}
                      onChange={(e) => update({ listId: e.target.value })}
                    >
                      <option value="">Choose a list</option>
                      {options.data?.lists.map((list) => (
                        <option key={list.id} value={list.id}>
                          {list.name}
                        </option>
                      ))}
                    </select>
                  </Field>
                )}
                {trigger === "segment.entered" && (
                  <Field
                    label="Which segment?"
                    hint="Segments are checked every 5 minutes. Only people who newly match after you turn this on start."
                  >
                    <select
                      value={String(config.segmentId || "")}
                      onChange={(e) => update({ segmentId: e.target.value })}
                    >
                      <option value="">Choose a segment</option>
                      {segments.data?.data?.map((segment) => (
                        <option key={segment.id} value={segment.id}>
                          {segment.name}
                          {segment.isDynamic ? "" : " (static)"}
                        </option>
                      ))}
                    </select>
                  </Field>
                )}
                {trigger === "date.reached" && (
                  <DateTriggerSettings config={config} update={update} />
                )}
                {trigger === "event.received" && (
                  <Field
                    label="Event name"
                    hint="Send it with POST /api/v1/events. Each property is available to conditions and emails as {{event_<name>}}, e.g. {{event_plan}}."
                  >
                    <input
                      maxLength={100}
                      placeholder="trial_started"
                      value={String(config.eventName || "")}
                      onChange={(e) => update({ eventName: e.target.value })}
                    />
                  </Field>
                )}
                <p className={workspaceClassName("inspector-help")}>
                  Only matching events in your workspace will start this
                  workflow.
                </p>
                <Field
                  label="Contacts can enter"
                  hint="A contact is never in this workflow twice at the same time."
                >
                  <select
                    value={reentryMode}
                    onChange={(e) => {
                      setReentryMode(e.target.value);
                      setDirty(true);
                    }}
                  >
                    <option value="once">Only once, ever</option>
                    <option value="after_finish">
                      Again after they finish
                    </option>
                    <option value="after_days">
                      Again after a number of days
                    </option>
                  </select>
                </Field>
                {reentryMode === "after_days" && (
                  <Field
                    label="Days since they last entered"
                    hint="Between 1 and 3650 days."
                  >
                    <input
                      type="number"
                      min={1}
                      max={3650}
                      value={reentryDays}
                      onChange={(e) => {
                        setReentryDays(Number(e.target.value));
                        setDirty(true);
                      }}
                    />
                  </Field>
                )}
                <label className="flex items-start gap-2 text-sm">
                  <input
                    type="checkbox"
                    checked={exitOnGoal}
                    onChange={(e) => {
                      setExitOnGoal(e.target.checked);
                      setDirty(true);
                    }}
                  />
                  <span>
                    Leave when they reach a goal of this workflow
                    <small className="block text-muted-foreground">
                      Contacts always leave when they unsubscribe, bounce,
                      complain, are suppressed or deleted.
                    </small>
                  </span>
                </label>
                <AutomationWindowSetting
                  value={sendWindow}
                  onChange={(w) => {
                    setSendWindow(w);
                    setDirty(true);
                  }}
                />
              </>
            )}
            {kind === "EMAIL" && (
              <>
                <Field label="Email template">
                  <select
                    value={String(config.templateId || "")}
                    onChange={(e) => update({ templateId: e.target.value })}
                  >
                    <option value="">Choose a template</option>
                    {options.data?.templates.map((t) => (
                      <option key={t.id} value={t.id}>
                        {t.name}
                      </option>
                    ))}
                  </select>
                </Field>
                <GmailClipNotice
                  html={
                    options.data?.templates.find(
                      (t) => t.id === config.templateId,
                    )?.htmlBody
                  }
                />
                <Field label="Send from">
                  <select
                    value={String(config.smtpConfigId || "")}
                    onChange={(e) => update({ smtpConfigId: e.target.value })}
                  >
                    <option value="">Choose a sender</option>
                    {options.data?.senders.map((s) => (
                      <option key={s.id} value={s.id}>
                        {s.fromEmail}
                      </option>
                    ))}
                  </select>
                </Field>
                <Field
                  label="Subject override"
                  hint="Leave blank to use the template subject."
                >
                  <input
                    maxLength={200}
                    value={String(config.subject || "")}
                    onChange={(e) => update({ subject: e.target.value })}
                  />
                </Field>
              </>
            )}
            {kind === "EMAIL" && (
              <Field
                label="Sender postal address"
                hint="Required for marketing templates; included with an unsubscribe link."
              >
                <textarea
                  maxLength={500}
                  value={String(config.postalAddress || "")}
                  onChange={(e) => update({ postalAddress: e.target.value })}
                />
              </Field>
            )}
            {kind === "EMAIL" && (
              <>
                <label className="flex items-start gap-2 text-sm">
                  <input
                    type="checkbox"
                    className="mt-1 !w-4"
                    checked={config.trackEngagement !== false}
                    onChange={(e) =>
                      update({ trackEngagement: e.target.checked })
                    }
                  />
                  <span>
                    Track opens and link clicks
                    <span className="block text-xs text-muted-foreground">
                      Off sends this email without a tracking pixel or rewritten
                      links.
                    </span>
                  </span>
                </label>
                <label className="flex items-start gap-2 text-sm">
                  <input
                    className="!w-4"
                    type="checkbox"
                    checked={!!config.transactional}
                    onChange={(e) => update({ transactional: e.target.checked })}
                  />
                  <span>
                    Transactional
                    <small className="mt-1 block text-muted-foreground">
                      Sends straight away, ignoring the send window and
                      frequency cap.
                    </small>
                  </span>
                </label>
                {!config.transactional && (
                  <label className="flex items-start gap-2 text-sm">
                    <input
                      className="!w-4"
                      type="checkbox"
                      checked={!!config.skipIfCapped}
                      onChange={(e) =>
                        update({ skipIfCapped: e.target.checked })
                      }
                    />
                    <span>
                      Skip if capped
                      <small className="mt-1 block text-muted-foreground">
                        When the contact has reached the frequency cap, move
                        on without sending instead of waiting.
                      </small>
                    </span>
                  </label>
                )}
              </>
            )}
            {kind === "WAIT" && (
              <WaitSettings config={config} update={update} />
            )}
            {kind === "WAIT_UNTIL" && (
              <WaitUntilSettings config={config} update={update} />
            )}
            <ActionSettings
              kind={kind}
              config={config}
              update={update}
              options={options.data}
            />
            {kind === "CONDITION" && (
              <ConditionSettings
                config={config}
                update={update}
                form={
                  trigger === "form.completed"
                    ? forms.data?.find(
                        (form) =>
                          form.id ===
                          nodes.find((node) => node.data.kind === "START")?.data
                            .formId,
                      )
                    : undefined
                }
                variables={nodes
                  .filter((n) => n.data.kind === "SET_VARIABLE")
                  .map((n) => String(n.data.variable))}
              />
            )}
            {branchPaths[kind] && (
              <>
                {edges
                  .filter((e) => e.source === selected)
                  .map((e) => (
                    <Field
                      key={e.id}
                      label={`Path to ${labels[String(nodes.find((n) => n.id === e.target)?.data.kind)] || "step"}`}
                    >
                      <select
                        value={String(e.label || "")}
                        onChange={(event) => {
                          setEdges((es) =>
                            es.map((edge) =>
                              edge.id === e.id
                                ? { ...edge, label: event.target.value }
                                : edge,
                            ),
                          );
                          setDirty(true);
                        }}
                      >
                        <option value="">Choose path</option>
                        {kind === "CONDITION" ? (
                          <>
                            <option value="true">Condition is true</option>
                            <option value="false">Condition is false</option>
                          </>
                        ) : kind === "WAIT_UNTIL" ? (
                          <>
                            <option value="happened">It happened</option>
                            <option value="timeout">Time ran out</option>
                          </>
                        ) : (
                          <>
                            <option value="A">Path A</option>
                            <option value="B">Path B</option>
                          </>
                        )}
                      </select>
                    </Field>
                  ))}
              </>
            )}
            {kind === "EXIT" && (
              <p className={workspaceClassName("inspector-help")}>
                Contacts reaching this step finish their journey. Every path
                should have a clear ending.
              </p>
            )}
            {!["START", "EXIT"].includes(kind) && (
              <Button variant="outline" onClick={remove} className="mt-5">
                <Trash2 />
                Remove step
              </Button>
            )}
          </fieldset>
          <div className={workspaceClassName("inspector-footer")}>
            <ShieldCheck size={15} />
            <span>Only active, subscribed contacts can receive emails.</span>
          </div>
        </aside>
      </div>
      {runs && <RunHistory id={id} close={() => setRuns(false)} />}{" "}
      {test && <RunWorkflow id={id} close={() => setTest(false)} />}
      {versions && (
        <VersionList
          id={id}
          live={liveVersion}
          restore={restore}
          close={() => setVersions(false)}
        />
      )}
      <Modal
        open={publishing}
        onOpenChange={setPublishing}
        title={`Publish version ${liveVersion + 1}?`}
        description={`New contacts will start on it. Contacts already inside stay on the version they started on.`}
      >
        <label className="flex items-start gap-2 text-sm">
          <input
            type="checkbox"
            className="mt-1"
            checked={moveRunning}
            onChange={(e) => setMoveRunning(e.target.checked)}
          />
          <span>
            Move running contacts to the new version
            <small className="block text-muted-foreground">
              They continue from the same step. Contacts on a step you removed
              leave the workflow.
            </small>
          </span>
        </label>
        <div className={workspaceClassName("modal-actions")}>
          <Button variant="outline" onClick={() => setPublishing(false)}>
            Cancel
          </Button>
          <Button
            onClick={() => {
              setPublishing(false);
              action("publish");
            }}
          >
            Publish
          </Button>
        </div>
      </Modal>
      <Modal
        open={leave}
        onOpenChange={setLeave}
        title="Leave this draft?"
        description="Your unsaved changes will be lost."
      >
        <div className={workspaceClassName("modal-actions")}>
          <Button variant="outline" onClick={() => setLeave(false)}>
            Keep editing
          </Button>
          <Button onClick={close}>Discard changes</Button>
        </div>
      </Modal>
    </div>
  );
}
function VersionList({
  id,
  live,
  restore,
  close,
}: {
  id: string;
  live: number;
  restore: (number: number) => void;
  close: () => void;
}) {
  const q = useMarketingQuery<WorkflowVersion[]>(`automations/${id}/versions`);
  return (
    <Modal
      open
      onOpenChange={close}
      title="Published versions"
      description="Every published version is kept. Restoring one replaces the draft; publish it to make it live."
    >
      <QueryState loading={q.isLoading} error={q.error} />
      <table className={workspaceClassName("product-table")}>
        <thead>
          <tr>
            <th>Version</th>
            <th>Published</th>
            <th />
          </tr>
        </thead>
        <tbody>
          {(Array.isArray(q.data) ? q.data : []).map((v) => (
            <tr key={v.number}>
              <td>
                Version {v.number}{" "}
                {v.number === live && <Status value="PUBLISHED" />}
              </td>
              <td>
                {new Date(v.publishedAt).toLocaleString()}
                <small>{v.publishedBy || "Automatic"}</small>
              </td>
              <td>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => restore(v.number)}
                >
                  Restore as draft
                </Button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </Modal>
  );
}
const exitReasons: Record<string, string> = {
  unsubscribed: "Left: unsubscribed",
  bounced: "Left: email bounced",
  complained: "Left: marked as spam",
  suppressed: "Left: on the do-not-email list",
  contact_deleted: "Left: contact deleted",
  goal_reached: "Left: reached the goal",
};
function RunHistory({ id, close }: { id: string; close: () => void }) {
  const q = useMarketingQuery<any[]>(`automations/${id}/executions`);
  return (
    <Modal
      open
      onOpenChange={close}
      title="Workflow history"
      description="Recent runs, progress, and errors for this workflow."
    >
      <QueryState loading={q.isLoading} error={q.error} />
      {q.data?.length === 0 && (
        <Empty
          title="No runs yet"
          description="Matching events will appear here after the workflow is published."
        />
      )}
      <table className={workspaceClassName("product-table")}>
        <thead>
          <tr>
            <th>Contact</th>
            <th>Status</th>
            <th>Started</th>
          </tr>
        </thead>
        <tbody>
          {q.data?.map((run) => (
            <tr key={run.id}>
              <td>
                {run.contact?.email || run.contactId}
                <small>{run.error}</small>
                {Array.isArray(run.executionLog) &&
                  run.executionLog.length > 0 && (
                    <details className="mt-2 text-xs">
                      <summary className="cursor-pointer">
                        {run.executionLog.length} step events
                      </summary>
                      <ol className="mt-3 space-y-3">
                        {run.executionLog.map((entry: any, index: number) => (
                          <li key={index} className="border-l-2 pl-3">
                            <strong>
                              {labels[entry.nodeType] || entry.nodeType}
                            </strong>{" "}
                            · {entry.status}
                            <p>{entry.message}</p>
                            {entry.error && (
                              <p className="text-destructive">{entry.error}</p>
                            )}
                            <time className="text-muted-foreground">
                              {new Date(entry.timestamp).toLocaleString()}
                            </time>
                          </li>
                        ))}
                      </ol>
                    </details>
                  )}
              </td>
              <td>
                <Status value={run.status} />
                {run.exitReason && (
                  <small>{exitReasons[run.exitReason] || run.exitReason}</small>
                )}
              </td>
              <td>{new Date(run.startedAt).toLocaleString()}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </Modal>
  );
}
function RunWorkflow({ id, close }: { id: string; close: () => void }) {
  const q = useMarketingQuery<Contact[]>("marketing/contacts");
  const { request } = useMarketing();
  const [contact, setContact] = useState("");
  const [busy, setBusy] = useState(false);
  return (
    <Modal
      open
      onOpenChange={close}
      title="Run this workflow"
      description="This executes the active workflow and can send real emails to the selected contact."
    >
      <Field label="Contact">
        <select value={contact} onChange={(e) => setContact(e.target.value)}>
          <option value="">Choose an active subscriber</option>
          {q.data
            ?.filter((c) => c.status === "ACTIVE")
            .map((c) => (
              <option value={c.id} key={c.id}>
                {c.email}
              </option>
            ))}
        </select>
      </Field>
      <Button
        className={workspaceClassName("product-primary")}
        disabled={!contact || busy}
        onClick={async () => {
          setBusy(true);
          try {
            const result = await request<any>(
              `automations/${id}/trigger`,
              "POST",
              { contactIds: [contact] },
            );
            if (!result.tasksEnqueued)
              throw new Error("The run could not be queued");
            toast.success("Workflow queued");
            close();
          } catch (e) {
            toast.error((e as Error).message);
          } finally {
            setBusy(false);
          }
        }}
      >
        {busy ? "Queuing…" : "Run for this contact"}
      </Button>
    </Modal>
  );
}
