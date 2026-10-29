"use client";

import { useRef, useState } from "react";
import { Button } from "@/components/ui/button";
import { Upload } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { parse, unparse } from "papaparse";
import { Label } from "@/components/ui/label";
import { Textarea } from "../ui/textarea";
import { toast } from "sonner";
import { useApi } from "@/hooks/use-api";
import {
  CONTACT_FIELDS,
  detectSource,
  guessField,
  prepareImport,
  type Row,
  type Source,
} from "@/lib/contacts/import-sources";

interface ContactImportProps {
  listId: string;
  onImportComplete: () => void;
}

/** CSV column → contact field, the shape the import worker expects. */
type FieldMapping = Record<string, string>;

const SOURCES = ["beehiiv", "Substack", "Mailchimp", "Ghost", "Kit"];
const LABELS: Record<string, string> = {
  email: "Email",
  first_name: "First name",
  last_name: "Last name",
  zip: "Postal code",
  linkedin: "LinkedIn",
};
const label = (f: string) =>
  LABELS[f] ?? f.charAt(0).toUpperCase() + f.slice(1);

export function ContactImport({
  listId,
  onImportComplete,
}: ContactImportProps) {
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [step, setStep] = useState<"upload" | "map" | "text">("upload");
  const [csvHeaders, setCsvHeaders] = useState<string[]>([]);
  const [fieldMapping, setFieldMapping] = useState<FieldMapping>({});
  const [rows, setRows] = useState<Row[]>([]);
  const [fileName, setFileName] = useState("contacts.csv");
  const [source, setSource] = useState<Source | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const { apiFetch } = useApi();
  const [text, setText] = useState(
    "email,first_name,last_name\njane@example.com,Jane,Smith",
  );

  const emailHeader = Object.keys(fieldMapping).find(
    (h) => fieldMapping[h] === "email",
  );
  const prepared =
    emailHeader && rows.length ? prepareImport(rows, emailHeader) : null;
  const skippedTotal = prepared
    ? Object.values(prepared.skipped).reduce((a, b) => a + b, 0)
    : 0;

  function load(input: File | string, name: string) {
    // Mailchimp exports unsubscribed and cleaned people as separate files.
    if (/unsubscribed|cleaned/i.test(name)) {
      toast.error(
        "This looks like a list of unsubscribed or cleaned contacts. Import your subscribed members file instead.",
      );
      return;
    }
    parse<Row>(input, {
      header: true,
      skipEmptyLines: true,
      transformHeader: (h) => h.trim(),
      complete: (results) => {
        if (results.data.length === 0) {
          toast.error("The CSV file is empty");
          return;
        }
        const headers = (results.meta.fields ?? []).filter(Boolean);
        const mapping: FieldMapping = {};
        for (const h of headers) {
          const field = guessField(h);
          if (field && !Object.values(mapping).includes(field))
            mapping[h] = field;
        }
        setFileName(name);
        setSource(detectSource(headers, name));
        setCsvHeaders(headers);
        setRows(results.data);
        setFieldMapping(mapping);
        setStep("map");
      },
      error: () => toast.error("Failed to parse CSV file"),
    });
  }

  async function handleImport() {
    if (!prepared || !emailHeader) {
      toast.error("Choose the column that holds email addresses");
      return;
    }
    if (prepared.keep.length === 0) {
      toast.error("No subscribed contacts with a valid email to import");
      return;
    }
    // The export only had "name"; prepareImport added these columns.
    const mappings = { ...fieldMapping };
    if ("first_name" in prepared.keep[0] && !Object.values(mappings).includes("first_name")) {
      mappings.first_name = "first_name";
      mappings.last_name = "last_name";
    }
    setIsLoading(true);
    try {
      const formData = new FormData();
      formData.append(
        "file",
        new File([unparse(prepared.keep)], fileName, { type: "text/csv" }),
      );
      const upload = await apiFetch("files/upload", {
        method: "POST",
        body: formData,
      });
      // files/upload responds with { file: "<id>" }.
      const fileId = upload.ok ? (await upload.json()).file : null;
      if (!fileId) throw new Error("the file could not be uploaded");
      const response = await apiFetch("imports/contact", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ mappings, listId, fileId }),
      });
      if (!response.ok) throw new Error(response.statusText);
      toast.success(
        `Importing ${prepared.keep.length.toLocaleString()} contacts. They’ll appear in this list shortly.`,
      );
      reset();
      setIsDialogOpen(false);
      onImportComplete();
    } catch (error) {
      toast.error(`Failed to import contacts: ${(error as Error).message}`);
    } finally {
      setIsLoading(false);
    }
  }

  function reset() {
    setStep("upload");
    setCsvHeaders([]);
    setFieldMapping({});
    setRows([]);
    setSource(null);
    if (fileInputRef.current) fileInputRef.current.value = "";
  }

  return (
    <Dialog
      open={isDialogOpen}
      onOpenChange={(open) => {
        setIsDialogOpen(open);
        if (!open) reset();
      }}
    >
      <DialogTrigger asChild>
        <Button>
          <Upload className="h-4 w-4 mr-2" />
          Import Contacts
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-[600px]">
        <DialogHeader>
          <DialogTitle>Import Contacts</DialogTitle>
          <div className="flex border-b py-4 gap-4">
            <Button
              variant={step !== "text" ? "secondary" : "ghost"}
              onClick={reset}
              size="sm"
            >
              CSV Import
            </Button>
            <Button
              variant={step === "text" ? "secondary" : "ghost"}
              onClick={() => setStep("text")}
              size="sm"
            >
              Paste contacts
            </Button>
          </div>
          <DialogDescription>
            {step === "upload" &&
              `Upload a CSV, or the subscriber export from ${SOURCES.slice(0, -1).join(", ")} or ${SOURCES.at(-1)}.`}
            {step === "map" && "Check which column holds each field."}
            {step === "text" && "Paste comma-separated contacts with a header row."}
          </DialogDescription>
        </DialogHeader>

        {step === "text" && (
          <div className="space-y-4 py-4">
            <Textarea
              value={text}
              onChange={(e) => setText(e.target.value)}
              aria-label="Contacts as CSV"
              rows={6}
              className="w-full font-mono text-xs"
            />
          </div>
        )}

        {step === "upload" && (
          <div className="space-y-4 py-4">
            <input
              ref={fileInputRef}
              type="file"
              accept=".csv,text/csv"
              aria-label="Contacts CSV file"
              onChange={(e) => {
                const file = e.target.files?.[0];
                if (file) load(file, file.name);
              }}
              className="w-full"
            />
            <ul className="list-disc list-inside text-sm text-muted-foreground">
              <li>An email column is required; other columns are kept as contact data.</li>
              <li>People who unsubscribed, bounced or were marked inactive in the export are left out.</li>
              <li>Duplicate and invalid addresses are left out.</li>
            </ul>
          </div>
        )}

        {step === "map" && (
          <div className="space-y-4 py-4 overflow-y-auto max-h-[500px]">
            {source && (
              <p className="rounded-md border px-3 py-2 text-sm">
                Looks like a <strong>{source}</strong> export. We matched its
                columns for you.
              </p>
            )}
            <div className="grid gap-3">
              {CONTACT_FIELDS.map((field) => (
                <div key={field} className="grid grid-cols-2 items-center gap-4">
                  <Label htmlFor={`map-${field}`}>
                    {label(field)}
                    {field === "email" && (
                      <span className="text-red-500 ml-1">*</span>
                    )}
                  </Label>
                  <Select
                    value={
                      Object.keys(fieldMapping).find(
                        (key) => fieldMapping[key] === field,
                      ) || "none"
                    }
                    onValueChange={(value) => {
                      const next = { ...fieldMapping };
                      for (const key of Object.keys(next))
                        if (next[key] === field) delete next[key];
                      if (value !== "none") next[value] = field;
                      setFieldMapping(next);
                    }}
                  >
                    <SelectTrigger id={`map-${field}`}>
                      <SelectValue placeholder="Select a column" />
                    </SelectTrigger>
                    <SelectContent className="max-h-60 overflow-y-auto">
                      <SelectItem value="none">Not mapped</SelectItem>
                      {csvHeaders.map((header) => (
                        <SelectItem key={header} value={header}>
                          {header}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
              ))}
            </div>

            <div className="rounded-md bg-muted/40 px-3 py-2 text-sm" aria-live="polite">
              {prepared ? (
                <>
                  <p>
                    <strong>{prepared.keep.length.toLocaleString()}</strong> of{" "}
                    {rows.length.toLocaleString()} rows will be imported.
                  </p>
                  {skippedTotal > 0 && (
                    <p className="mt-1 text-muted-foreground">
                      {skippedTotal.toLocaleString()} left out:{" "}
                      {Object.entries(prepared.skipped)
                        .map(([why, n]) => `${n.toLocaleString()} ${why}`)
                        .join(", ")}
                      .
                    </p>
                  )}
                </>
              ) : (
                <p>Choose the column that holds email addresses.</p>
              )}
            </div>
          </div>
        )}

        <DialogFooter>
          <Button
            variant="outline"
            onClick={() => {
              reset();
              setIsDialogOpen(false);
            }}
          >
            Cancel
          </Button>
          {step === "text" && (
            <Button onClick={() => load(text, "pasted-contacts.csv")}>
              Continue
            </Button>
          )}
          {step === "map" && (
            <Button
              onClick={handleImport}
              disabled={isLoading || !prepared?.keep.length}
            >
              {isLoading
                ? "Importing..."
                : `Import ${prepared?.keep.length.toLocaleString() ?? ""} contacts`}
            </Button>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
