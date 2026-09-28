export interface Form {
  id: string;
  name: string;
  description?: string;
  createdAt: Date;
  updatedAt: Date;
  mailingListId?: string;
  teamId: string;
  status: "draft" | "active" | "archived";
  fields: FormField[];
}

export interface FormField {
  id: string;
  label: string;
  type: "text" | "email" | "number" | "textarea" | "select" | "checkbox";
  required: boolean;
  placeholder?: string;
  options?: string[]; // For select fields
  defaultValue?: string;
}

export interface Attachment {
  id?: string;
  name: string;
  data: string;
}

export interface Mail {
  id: string;
  subject: string;
  body: string;
  from: string;
  to: string[];
  cc?: string[];
  bcc?: string[];
  replyTo?: string;
  templateId?: string;
  campaignId?: string;
  contactId?: string;
  categoryId: string;
  smtpConfigId: string;
  teamId: string;
  status: "draft" | "scheduled" | "sending" | "sent" | "failed";
  sentAt?: Date;
  createdAt: Date;
  updatedAt: Date;
  isDeleted?: boolean;
  error?: string;
  data?: Record<string, any>;
  test?: boolean;
  attachments?: Attachment[];
}
