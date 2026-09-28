export interface Contact {
  id: string;
  email: string;
  firstName: string | null;
  lastName: string | null;
  status: string;
  createdAt: string;
}

export interface MailingList {
  id: string;
  name: string;
  description: string | null;
  createdAt: string;
  _count: {
    subscribers: number;
  };
} 