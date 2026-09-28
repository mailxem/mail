"use client";
import { ContactsList } from "@/components/contacts/contacts-list";
import { useParams } from "next/navigation";
export default function ContactListPage() {
  const { listId } = useParams<{ listId: string }>();
  return <ContactsList key={listId} listId={listId}/>;
}
