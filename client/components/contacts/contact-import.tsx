"use client";

import { useState, useRef, useEffect } from "react";
import { Button } from "@/components/ui/button";
import { Upload } from "lucide-react";
import { useTeam } from "@/app/providers/team-provider";
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
import { parse } from "papaparse";
import { Label } from "@/components/ui/label";
import { Textarea } from "../ui/textarea";
import { toast } from "sonner";
import { useApi } from "@/hooks/use-api";

interface ContactImportProps {
  listId: string;
  onImportComplete: () => void;
}

interface FieldMapping {
  [key: string]: string;
}

// Define field mappings for common CSV headers
const FIELD_MAPPINGS: { [key: string]: string[] } = {
  email: ["email", "e-mail", "e_mail", "email_address", "emailaddress"],
  firstName: ["first_name", "firstname", "first", "given_name", "givenname"],
  lastName: ["last_name", "lastname", "last", "surname", "family_name"],
  phone: ["phone", "phone_number", "phonenumber", "mobile", "telephone", "tel"],
  company: [
    "company",
    "organization",
    "organisation",
    "company_name",
    "employer",
  ],
  title: ["title", "job_title", "position", "role"],
  address: ["address", "address_line_1", "street_address"],
  city: ["city", "town"],
  state: ["state", "province", "region", "state_province_region"],
  country: ["country", "country_name"],
  postalCode: ["postal_code", "zip", "zip_code", "postcode"],
};

const REQUIRED_FIELDS = ["email"];
const OPTIONAL_FIELDS = [
  "first_name",
  "last_name",
  "phone",
  "company",
  "title",
  "address",
  "city",
  "state",
  "country",
  "postal_code",
  "linkedin",
];
const ALL_FIELDS = [...REQUIRED_FIELDS, ...OPTIONAL_FIELDS];

export function ContactImport({
  listId,
  onImportComplete,
}: ContactImportProps) {
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [step, setStep] = useState<"upload" | "map" | "preview" | "text">(
    "upload"
  );
  const [csvHeaders, setCsvHeaders] = useState<string[]>([]);
  const [fieldMapping, setFieldMapping] = useState<FieldMapping>({});
  const [csvData, setCsvData] = useState<any[]>([]);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const { team } = useTeam();
  const { apiFetch } = useApi();
  const [file, setFile] = useState<File | null>(null);
  const [fileId, setFileId] = useState<string | null>(null);
  const [text, setText] = useState<string>(
    "name,email,phone\nJohn Doe,john.doe@example.com,1234567890\nJane Smith,jane.smith@example.com,9876543210"
  );
  // Function to guess field mapping based on header name
  const guessFieldMapping = (header: string): string | null => {
    const normalizedHeader = header.toLowerCase().replace(/[^a-z0-9]/g, "");

    for (const [field, variations] of Object.entries(FIELD_MAPPINGS)) {
      if (
        variations.some((v) =>
          normalizedHeader.includes(v.replace(/[^a-z0-9]/g, ""))
        )
      ) {
        return field;
      }
    }
    return null;
  };

  const handleFileUpload = async (
    event: React.ChangeEvent<HTMLInputElement>
  ) => {
    const file = event.target.files?.[0];
    if (!file) return;

    setFile(file);

    parse(file, {
      header: true,
      skipEmptyLines: true,
      transformHeader: (header) => {
        return header.trim();
      },
      complete: (results) => {
        if (results.data.length === 0) {
          toast.error("The CSV file is empty");
          return;
        }

        // Get headers from the first row
        const headers = Object.keys(results.data[0]);
        setCsvHeaders(headers.filter((header) => header !== ""));
        setCsvData(results.data);

        // Initialize field mapping with best guesses
        const initialMapping: FieldMapping = {};
        headers.forEach((header) => {
          const guessedField = guessFieldMapping(header);
          if (guessedField) {
            initialMapping[header] = guessedField;
          }
        });
        setFieldMapping(initialMapping);
        setStep("map");
      },
      error: () => {
        toast.error("Failed to parse CSV file");
      },
    });
  };

  const handleImport = async () => {
    if (step === "text") {
      toast.loading("Creating CSV file...");
      // lets create a csv file from the text
      const csv = text
        .split("\n")
        .map((line) =>
          line
            .split(",")
            .map((cell) => cell.trim())
            .join(",")
        )
        .join("\n");
      setFile(new File([csv], "contacts.csv", { type: "text/csv" }));
      fieldMapping["name"] = "name";
      fieldMapping["email"] = "email";
      fieldMapping["phone"] = "phone";
    }

    // Validate required fields
    if (!Object.values(fieldMapping).includes("email")) {
      toast.dismiss();
      toast.error("Email field mapping is required");
      return;
    }

    try {
      setIsLoading(true);

      if (!fileId) {
        // upload file
        // form data
        const formData = new FormData();
        formData.append("file", file);
        const fileUpload = await apiFetch("files/upload", {
          method: "POST",
          body: formData,
        });

        const fileUploadResponse = await fileUpload.json();

        setFileId(fileUploadResponse.data.fileId);
      }

      if (!fileId) {
        toast.dismiss();
        toast.error("Failed to upload file");
        return;
      }

      const response = await apiFetch("imports/contact", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          mappings: fieldMapping,
          listId,
          teamId: team?.id,
          fileId: fileId || "",
        }),
      });

      if (!response.ok) {
        throw new Error("Failed to import contacts: " + response.statusText);
      }

      const result = await response.json();
      toast.success(`Successfully imported ${result.imported} contacts`);

      // Reset state and close dialog
      setStep("upload");
      setFileId(null);
      setFile(null);
      setText("");
      setCsvHeaders([]);
      setFieldMapping({});
      setCsvData([]);
      if (fileInputRef.current) {
        fileInputRef.current.value = "";
      }
      setIsDialogOpen(false);
      onImportComplete();
    } catch (error) {
      toast.dismiss();
      console.error("Failed to import contacts", error);
      toast.error("Failed to import contacts: " + error.message);
    } finally {
      setIsLoading(false);
      toast.dismiss();
      if (fileInputRef.current) {
        fileInputRef.current.value = "";
      }
    }
  };

  const handleCancel = () => {
    setStep("upload");
    setIsDialogOpen(false);
    setCsvHeaders([]);
    setFieldMapping({});
    setCsvData([]);
    if (fileInputRef.current) {
      fileInputRef.current.value = "";
    }
  };

  return (
    <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
      <DialogTrigger asChild>
        <Button>
          <Upload className="h-4 w-4 mr-2" />
          Import Contacts
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-[600px]">
        <DialogHeader>
          <DialogTitle>Import Contacts</DialogTitle>
          <DialogDescription>
            <div className="flex border-b py-4 mb-3 gap-4">
              <Button
                variant={step === "upload" ? "secondary" : "ghost"}
                onClick={() => setStep("upload")}
                size="sm"
              >
                CSV Import
              </Button>
              <Button
                variant={step === "text" ? "secondary" : "ghost"}
                onClick={() => setStep("text")}
                size="sm"
              >
                Manually Add
              </Button>
            </div>
            {step === "upload" && "Upload a CSV file with your contacts data."}
            {step === "map" && "Map your CSV columns to contact fields."}
            {step === "text" && "Manually add contacts."}
          </DialogDescription>
        </DialogHeader>

        {step === "text" && (
          <div className="space-y-4 py-4">
            <Textarea
              value={text}
              onChange={(e) => setText(e.target.value)}
              placeholder="Enter contacts here"
              className="w-full"
            />
          </div>
        )}

        {step === "upload" && (
          <div className="space-y-4 py-4">
            <input
              ref={fileInputRef}
              type="file"
              accept=".csv"
              onChange={handleFileUpload}
              className="w-full"
            />
            <div className="text-sm text-muted-foreground">
              <p>Your CSV file should include the following:</p>
              <ul className="list-disc list-inside mt-2">
                <li>Required: email</li>
                <li>
                  Optional: firstName, lastName, phone, company, title, address,
                  city, state, country, postalCode
                </li>
                <li>Any additional fields will be stored as metadata</li>
              </ul>
            </div>
          </div>
        )}

        {step === "map" && (
          <div className="space-y-4 py-4 overflow-y-auto max-h-[500px]">
            <div className="grid gap-4">
              {ALL_FIELDS.map((field) => (
                <div
                  key={field}
                  className="grid grid-cols-2 items-center gap-4"
                >
                  <Label className="capitalize">
                    {field}
                    {REQUIRED_FIELDS.includes(field) && (
                      <span className="text-red-500 ml-1">*</span>
                    )}
                  </Label>
                  <Select
                    value={
                      Object.keys(fieldMapping)?.find(
                        (key) => fieldMapping[key] === field
                      ) || "none"
                    }
                    key={field}
                    onValueChange={(value) => {
                      const newMapping = { ...fieldMapping };
                      // Remove old mapping if it exists
                      Object.keys(newMapping).forEach((key) => {
                        if (newMapping[key] === field) {
                          delete newMapping[key];
                        }
                      });
                      // Add new mapping if a column is selected
                      if (value !== "none") {
                        newMapping[value] = field;
                      }
                      setFieldMapping(newMapping);
                    }}
                  >
                    <SelectTrigger className="text-muted-foreground">
                      <SelectValue
                        className="capitalize"
                        placeholder="Select a column"
                      />
                    </SelectTrigger>
                    <SelectContent className="h-54 overflow-y-auto">
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

            <div className="text-sm text-muted-foreground mt-4">
              <p>Preview: {csvData.length} contacts will be imported</p>
              <p className="mt-1">
                Unmapped columns will be stored as metadata
              </p>
            </div>
          </div>
        )}

        <DialogFooter>
          <Button variant="outline" onClick={handleCancel}>
            Cancel
          </Button>
          {step === "map" ||
            (step === "text" && (
              <Button onClick={handleImport} disabled={isLoading}>
                {isLoading ? "Importing..." : "Import Contacts"}
              </Button>
            ))}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
