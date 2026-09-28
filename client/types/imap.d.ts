export interface IMAPEmail {
  id?: string;
  uid?: number;
  uidValidity?: number;
  reply_to?: string;
  subject: string;
  from: string;
  to: string;
  cc: string;
  bcc: string;
  body: string;
  date: string;
  messageId: string;
  flags: string[] | null;
  attachments: {
    Filename: string;
    Data: string;
  }[];
}

export interface IMAPEmailResponse {
  emails: IMAPEmail[];
  total_emails: number;
  limit: number;
  offset: number;
}
