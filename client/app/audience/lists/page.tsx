"use client";

import { MailingListProvider } from "@/app/providers/mailinglist-provider";
import { PageHeader } from "@/components/page-header";
import { ContactsListHeader } from "@/components/contacts/contacts-list-header";
import { ContactLists } from "@/components/contacts/contact-lists";

export default function ContactListsPage() {
  return (
    <MailingListProvider>
      <PageHeader heading="Contact lists" description="The audiences you use across campaigns, newsletters, and forms."><ContactsListHeader/></PageHeader>
      <ContactLists />
    </MailingListProvider>
  );
}
