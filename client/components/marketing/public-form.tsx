"use client";

import { useEffect, useMemo, useRef, useState, type FormEvent } from "react";
import { Check, Loader2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { workspaceClassName } from "@/lib/workspace-styles";
import type { FormField } from "@/lib/marketing/types";
import { publicFormAction, type FormTheme } from "@/lib/marketing/form-theme";
import {
  getFormDefinition,
  getReachableAnswers,
  getReachablePages,
  validateFormPage,
  type FormAnswers,
  type FormDefinition,
  type FormQuestion,
} from "@/lib/marketing/form-definition";
import { FormSurface } from "./form-surface";
import styles from "./public-form.module.css";

type PublicFormData = {
  theme?: Partial<FormTheme>;
  name: string;
  description: string;
  fields: FormField[];
  definition?: FormDefinition;
  version?: number;
  buttonText: string;
  successMessage: string;
};
type PendingRequest = {
  requestId: string;
  fields: FormAnswers;
  consent: boolean;
  website: string;
  version?: number;
  sessionId: string;
  resumeToken?: string;
  pageId?: string;
  attribution: Record<string, string>;
};

export function PublicForm({ slug }: { slug: string }) {
  return <FormRuntime key={slug} slug={slug} />;
}

function FormRuntime({ slug }: { slug: string }) {
  const [form, setForm] = useState<PublicFormData | null>(null);
  const [answers, setAnswers] = useState<FormAnswers>({});
  const [consent, setConsent] = useState(false);
  const [pageId, setPageId] = useState("");
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [error, setError] = useState("");
  const [notice, setNotice] = useState("");
  const [busy, setBusy] = useState<"submit" | "save" | null>(null);
  const [done, setDone] = useState(false);
  const [embedded, setEmbedded] = useState(false);
  const [resumeToken, setResumeToken] = useState<string>();
  const [savedLink, setSavedLink] = useState("");
  const [expiresAt, setExpiresAt] = useState("");
  const session = useRef("");
  const initialResumeToken = useRef<string | null>(null);
  const started = useRef(false);
  const tracked = useRef(new Set<string>());
  const pendingSubmission = useRef<PendingRequest | null>(null);
  const pendingSave = useRef<PendingRequest | null>(null);
  const attribution = useRef<Record<string, string>>({});
  const heading = useRef<HTMLHeadingElement>(null);
  const element = useRef<HTMLFormElement>(null);
  const website = useRef<HTMLInputElement>(null);
  const definition = useMemo(
    () => (form ? getFormDefinition(form) : null),
    [form],
  );
  const pages = useMemo(
    () => (definition ? getReachablePages(definition, answers) : []),
    [definition, answers],
  );
  const pageIndex = Math.max(
    0,
    pages.findIndex((page) => page.id === pageId),
  );
  const page = pages[pageIndex];
  const lastPage = pageIndex === pages.length - 1;

  useEffect(() => {
    const controller = new AbortController();
    const params = new URLSearchParams(window.location.search);
    for (const [query, key] of [
      ["utm_source", "utmSource"],
      ["utm_medium", "utmMedium"],
      ["utm_campaign", "utmCampaign"],
      ["utm_content", "utmContent"],
      ["utm_term", "utmTerm"],
    ]) {
      const value = params.get(query);
      if (value) attribution.current[key] = value.slice(0, 200);
    }
    try {
      if (document.referrer)
        attribution.current.referrer = new URL(document.referrer).origin;
    } catch {
      /* No referrer is fine. */
    }
    const storageKey = `xem-form-session:${slug}`;
    session.current = crypto.randomUUID();
    try {
      const stored = sessionStorage.getItem(storageKey);
      if (stored && /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i.test(stored))
        session.current = stored;
      else sessionStorage.setItem(storageKey, session.current);
    } catch {
      /* Form use does not require browser storage. */
    }
    const fragment = new URLSearchParams(window.location.hash.slice(1));
    initialResumeToken.current ||= fragment.get("resume");
    const token = initialResumeToken.current;
    if (token) {
      fragment.delete("resume");
      history.replaceState(
        null,
        "",
        `${location.pathname}${location.search}${fragment.size ? `#${fragment}` : ""}`,
      );
    }
    async function load() {
      try {
        const response = await fetch(publicFormAction(slug), {
          signal: controller.signal,
        });
        if (!response.ok)
          throw new Error("This form is not available right now.");
        const data: PublicFormData = await response.json();
        if (token) {
          try {
            const resumed = await fetch(`${publicFormAction(slug)}/resume`, {
              method: "POST",
              headers: { "Content-Type": "application/json" },
              body: JSON.stringify({ token }),
              signal: controller.signal,
            });
            const body = await resumed.json();
            if (!resumed.ok)
              throw new Error(
                body.error ||
                  "This saved link has expired. You can start again below.",
              );
            setAnswers(body.fields || {});
            setPageId(body.pageId || "");
            setResumeToken(token);
            if (body.sessionId) {
              session.current = body.sessionId;
              try {
                sessionStorage.setItem(storageKey, body.sessionId);
              } catch {}
            }
            setNotice(
              body.version !== data.version
                ? "This form has changed since you saved it. Your answers are still here; please review each step."
                : "Your saved answers are ready. Please review them before submitting.",
            );
            if (body.version !== data.version) setPageId("");
          } catch (e) {
            if (controller.signal.aborted) return;
            setNotice((e as Error).message);
          }
        }
        if (!controller.signal.aborted) setForm(data);
      } catch (e) {
        if (!controller.signal.aborted) setError((e as Error).message);
      }
    }
    void load();
    return () => controller.abort();
  }, [slug]);

  function track(event: "view" | "start" | "step", step?: string) {
    if (!form || !session.current) return;
    const key = `${session.current}:${form.version}:${event}:${step || ""}`;
    if (tracked.current.has(key)) return;
    tracked.current.add(key);
    void fetch(`${publicFormAction(slug)}/events`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        sessionId: session.current,
        event,
        pageId: step,
        version: form.version,
        website: "",
      }),
      keepalive: true,
    }).then((response) => {
      if (!response.ok) tracked.current.delete(key);
    }).catch(() => tracked.current.delete(key));
  }
  useEffect(() => {
    if (form) track("view");
  }, [form]); // eslint-disable-line react-hooks/exhaustive-deps
  useEffect(() => {
    if (page) track("step", page.id);
  }, [page?.id, form]); // eslint-disable-line react-hooks/exhaustive-deps

  // Only send non-sensitive sizing/completion signals to the explicitly named parent origin.
  useEffect(() => {
    if (window.parent === window) return;
    setEmbedded(true);
    const parentOrigin = new URLSearchParams(location.search).get(
      "embedOrigin",
    );
    if (!parentOrigin) return;
    try {
      const url = new URL(parentOrigin);
      if (
        !["https:", "http:"].includes(url.protocol) ||
        url.origin !== parentOrigin
      )
        return;
    } catch {
      return;
    }
    const surface = document.querySelector<HTMLElement>("[data-form-surface]") || document.body;
    const sendSize = () =>
      window.parent.postMessage(
        { type: "xem:resize", height: Math.ceil(surface.getBoundingClientRect().height) },
        parentOrigin,
      );
    const observer = new ResizeObserver(sendSize);
    observer.observe(surface);
    sendSize();
    if (done) window.parent.postMessage({ type: "xem:complete" }, parentOrigin);
    return () => observer.disconnect();
  }, [done]);

  function change(key: string, value: string) {
    if (!started.current) {
      started.current = true;
      track("start");
    }
    setAnswers((previous) => ({ ...previous, [key]: value }));
    setErrors((previous) => {
      const next = { ...previous };
      delete next[key];
      return next;
    });
    setSavedLink("");
    // A known unsuccessful request can be edited. An uncertain completion stays immutable for retry.
    pendingSave.current = null;
  }
  function goTo(id: string) {
    setPageId(id);
    setErrors({});
    setError("");
    requestAnimationFrame(() => heading.current?.focus());
  }
  function checkPage(): boolean {
    if (!page || !definition) return false;
    const result = validateFormPage(
      page,
      getReachableAnswers(definition, answers),
    );
    setErrors(result);
    if (Object.keys(result).length) {
      requestAnimationFrame(() =>
        element.current
          ?.querySelector<HTMLElement>('input[aria-invalid="true"], select[aria-invalid="true"], textarea[aria-invalid="true"]')
          ?.focus(),
      );
      return false;
    }
    return true;
  }
  function payload(): PendingRequest {
    return {
      requestId: crypto.randomUUID(),
      fields: getReachableAnswers(definition!, answers),
      consent: definition?.consent.mode !== "none" && consent,
      website: website.current?.value || "",
      version: form?.version,
      sessionId: session.current,
      resumeToken,
      pageId: page?.id,
      attribution: attribution.current,
    };
  }
  async function refreshRevision() {
    const response = await fetch(publicFormAction(slug));
    if (!response.ok)
      throw new Error(
        "The form has changed, but its latest version could not be loaded. Your answers are still here. Please try again.",
      );
    setForm(await response.json());
    setPageId("");
    setSavedLink("");
    setResumeToken(undefined);
    setExpiresAt("");
    setConsent(false);
    pendingSave.current = null;
    setNotice(
      "This form has changed. Your answers are still here; please review each step and submit again.",
    );
  }
  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (busy || !definition) return;
    if (!started.current) {
      started.current = true;
      track("start");
    }
    if (!pendingSubmission.current) {
      if (!checkPage()) return;
      if (!lastPage) {
        goTo(pages[pageIndex + 1].id);
        return;
      }
      for (const candidate of pages) {
        const result = validateFormPage(
          candidate,
          getReachableAnswers(definition, answers),
        );
        if (Object.keys(result).length) {
          setPageId(candidate.id);
          setErrors(result);
          requestAnimationFrame(() => element.current?.querySelector<HTMLElement>('input[aria-invalid="true"], select[aria-invalid="true"], textarea[aria-invalid="true"]')?.focus());
          return;
        }
      }
      if (definition.consent.mode === "required" && !consent) {
        setError("Please agree to receive emails before submitting.");
        element.current?.querySelector<HTMLInputElement>('input[name="consent"]')?.focus();
        return;
      }
      pendingSubmission.current = payload();
    }
    setError("");
    setBusy("submit");
    try {
      const response = await fetch(publicFormAction(slug), {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(pendingSubmission.current),
      });
      const body = await response.json();
      if (!response.ok) {
        if (
          response.status === 409 &&
          (body.code === "FORM_VERSION_CHANGED" ||
            body.code === "VERSION_CONFLICT" ||
            /version|changed|revision/i.test(body.error || body.message || ""))
        ) {
          pendingSubmission.current = null;
          await refreshRevision();
          return;
        }
        if (response.status < 500) pendingSubmission.current = null;
        throw new Error(
          body.error ||
            body.message ||
            "We couldn’t submit your form. Please try again.",
        );
      }
      setDone(true);
      pendingSubmission.current = null;
      try {
        sessionStorage.removeItem(`xem-form-session:${slug}`);
      } catch {}
      if (body.successRedirectUrl) {
        try {
          const url = new URL(body.successRedirectUrl);
          if (url.protocol === "https:" && !url.username && !url.password)
            window.location.assign(url.href);
        } catch {
          /* The confirmation remains visible for an invalid redirect. */
        }
      }
    } catch (e) {
      setError(
        `${(e as Error).message}${pendingSubmission.current ? " Retry to check this same submission safely." : ""}`,
      );
    } finally {
      setBusy(null);
    }
  }
  async function saveProgress() {
    if (busy || !definition?.saveProgress) return;
    setBusy("save");
    setError("");
    pendingSave.current ||= payload();
    try {
      const response = await fetch(`${publicFormAction(slug)}/progress`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(pendingSave.current),
      });
      const body = await response.json();
      if (!response.ok) {
        if (
          response.status === 409 &&
          /version|changed|revision/i.test(body.error || body.message || "")
        ) {
          pendingSave.current = null;
          await refreshRevision();
          return;
        }
        if (response.status < 500) pendingSave.current = null;
        throw new Error(
          body.error ||
            body.message ||
            "Your answers could not be saved. Please try again.",
        );
      }
      setResumeToken(body.resumeToken);
      // Fragments are not sent in HTTP requests or referrer headers.
      const link = new URL(location.pathname, location.origin);
      link.hash = new URLSearchParams({ resume: body.resumeToken }).toString();
      setSavedLink(link.href);
      setExpiresAt(body.expiresAt);
      pendingSave.current = null;
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setBusy(null);
    }
  }

  return (
    <FormSurface theme={form?.theme} name={form?.name} embedded={embedded}>
      {done ? (
        <div className={workspaceClassName("product-empty !p-4")} role="status">
          <Check size={32} />
          <h1>Thank you.</h1>
          <p>{form?.successMessage || "Your response has been received."}</p>
        </div>
      ) : form && definition ? (
        <>
          <h1>{form.name}</h1>
          {form.description && <p>{form.description}</p>}
          {notice && (
            <p role="status" className={styles.notice}>
              {notice}
            </p>
          )}
          {pages.length > 1 && (
            <div className={styles.progress}>
              <span>
                Step {pageIndex + 1} of {pages.length}
              </span>
              <progress
                aria-label="Form progress"
                value={pageIndex + 1}
                max={pages.length}
              />
            </div>
          )}
          <form
            ref={element}
            className={`${workspaceClassName("product-form")} ${styles.form}`}
            onSubmit={submit}
            noValidate
          >
            {page && (
              <>
                {pages.length > 1 && <h2 ref={heading} tabIndex={-1} className={styles.stepTitle}>{page.title}</h2>}
                {page.description && <p>{page.description}</p>}
                <fieldset
                  className={styles.fields}
                  disabled={!!busy || !!pendingSubmission.current}
                >
                  {page.fields.map((field) => (
                    <Question
                      key={field.key}
                      field={field}
                      value={answers[field.key] ?? field.defaultValue ?? ""}
                      error={errors[field.key]}
                      onChange={(value) => change(field.key, value)}
                    />
                  ))}
                  {lastPage && definition.consent.mode !== "none" && (
                    <label
                      className={`${workspaceClassName("consent-row")} ${styles.consent}`}
                    >
                      <input
                        type="checkbox"
                        name="consent"
                        required={definition.consent.mode === "required"}
                        checked={consent}
                        onChange={(e) => {
                          setConsent(e.target.checked);
                          pendingSave.current = null;
                        }}
                      />
                      <span>
                        {definition.consent.label}
                        {definition.consent.mode === "optional" && (
                          <small>
                            Optional. You can submit without subscribing.
                          </small>
                        )}
                      </span>
                    </label>
                  )}
                </fieldset>
              </>
            )}
            <div
              aria-hidden="true"
              style={{ position: "absolute", left: "-10000px" }}
            >
              <label>
                Website
                <input
                  ref={website}
                  name="website"
                  tabIndex={-1}
                  autoComplete="off"
                />
              </label>
            </div>
            {error && (
              <div className={workspaceClassName("product-error")} role="alert">
                {error}
              </div>
            )}
            <div className={styles.actions}>
              {pageIndex > 0 && (
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => goTo(pages[pageIndex - 1].id)}
                  disabled={!!busy || !!pendingSubmission.current}
                >
                  Back
                </Button>
              )}
              <Button type="submit" disabled={!!busy || !page}>
                {busy === "submit"
                  ? "Submitting…"
                  : pendingSubmission.current
                    ? "Retry submission"
                    : lastPage
                      ? form.buttonText || "Submit"
                      : "Continue"}
              </Button>
            </div>
            {definition.saveProgress && (
              <div className={styles.save}>
                <button
                  type="button"
                  onClick={saveProgress}
                  disabled={!!busy || !!pendingSubmission.current}
                >
                  {busy === "save" ? "Saving…" : "Save and continue later"}
                </button>
                <small>
                  Only choosing Save stores your unfinished answers.
                </small>
              </div>
            )}
            {savedLink && (
              <div className={styles.saved} role="status">
                <strong>Your progress is saved.</strong>
                <label htmlFor="resume-link">
                  Keep this private link to continue later.
                </label>
                <input
                  id="resume-link"
                  readOnly
                  value={savedLink}
                  onFocus={(e) => e.target.select()}
                />
                <button
                  type="button"
                  onClick={async () => {
                    try {
                      await navigator.clipboard.writeText(savedLink);
                      setNotice("Resume link copied.");
                    } catch {
                      setNotice("Select and copy the resume link below.");
                    }
                  }}
                >
                  Copy resume link
                </button>
                {expiresAt && (
                  <small>
                    Expires {new Date(expiresAt).toLocaleDateString()}. Anyone
                    with this link can access your saved answers.
                  </small>
                )}
              </div>
            )}
          </form>
        </>
      ) : error ? (
        <p role="alert">{error}</p>
      ) : (
        <div role="status" className={workspaceClassName("product-empty")}>
          <Loader2 className="animate-spin" />
          Loading form…
        </div>
      )}
      <p className="!mb-0 !mt-7 !text-center !text-[10px]">Powered by Xem</p>
    </FormSurface>
  );
}

function Question({
  field,
  value,
  error,
  onChange,
}: {
  field: FormQuestion;
  value: string;
  error?: string;
  onChange: (value: string) => void;
}) {
  if (field.type === "HIDDEN") return null;
  const id = `question-${field.key}`;
  const description =
    [field.helpText ? `${id}-help` : "", error ? `${id}-error` : ""]
      .filter(Boolean)
      .join(" ") || undefined;
  const shared = {
    id,
    name: field.key,
    required: field.required,
    "aria-invalid": error ? (true as const) : undefined,
    "aria-describedby": description,
  };
  const label = `${field.label}${field.required ? " *" : ""}`;
  return (
    <div className={styles.question}>
      {field.type === "RADIO" ? (
        <fieldset
          className={styles.radio}
          aria-describedby={description}
          aria-invalid={!!error}
        >
          <legend>{label}</legend>
          {field.options?.map((option, index) => (
            <label key={option}>
              <input
                {...shared}
                id={index === 0 ? id : `${id}-${index}`}
                type="radio"
                value={option}
                checked={value === option}
                onChange={() => onChange(option)}
              />
              {option}
            </label>
          ))}
        </fieldset>
      ) : field.type === "CHECKBOX" ? (
        <label className={styles.consent} htmlFor={id}>
          <input
            {...shared}
            type="checkbox"
            checked={value === "true"}
            onChange={(e) => onChange(e.target.checked ? "true" : "false")}
          />
          <span>{label}</span>
        </label>
      ) : (
        <>
          <label htmlFor={id}>{label}</label>
          {field.type === "TEXTAREA" ? (
            <textarea
              {...shared}
              placeholder={field.placeholder}
              maxLength={2000}
              value={value}
              onChange={(e) => onChange(e.target.value)}
              rows={4}
            />
          ) : field.type === "SELECT" ? (
            <select
              {...shared}
              value={value}
              onChange={(e) => onChange(e.target.value)}
            >
              <option value="">
                {field.placeholder || "Choose an option"}
              </option>
              {field.options?.map((option) => (
                <option key={option} value={option}>
                  {option}
                </option>
              ))}
            </select>
          ) : (
            <input
              {...shared}
              type={
                field.type === "EMAIL"
                  ? "email"
                  : field.type === "PHONE"
                    ? "tel"
                    : field.type === "NUMBER"
                      ? "number"
                      : field.type === "DATE"
                        ? "date"
                        : "text"
              }
              step={field.type === "NUMBER" ? "any" : undefined}
              autoComplete={
                field.key === "email"
                  ? "email"
                  : ["firstName", "first_name"].includes(field.key)
                    ? "given-name"
                    : ["lastName", "last_name"].includes(field.key)
                      ? "family-name"
                      : field.type === "PHONE"
                        ? "tel"
                        : undefined
              }
              placeholder={field.placeholder}
              maxLength={2000}
              value={value}
              onChange={(e) => onChange(e.target.value)}
            />
          )}
        </>
      )}
      {field.helpText && <small id={`${id}-help`}>{field.helpText}</small>}
      {error && (
        <span id={`${id}-error`} className={styles.fieldError}>
          {error}
        </span>
      )}
    </div>
  );
}
