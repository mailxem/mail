import { Mail } from "@/app/types";
import { IMAPEmail } from "@/types/imap";

export const parsedMailFrom = (mail: Mail | IMAPEmail) => {
  if (mail) {
    const emailRegex = /<([^>]+)>/;
    const match = mail.from.match(emailRegex);
    const name = mail.from.split(" ")[0];
    return match ? `${name.replace(/[^\w\s]/g, "")} <${match[1]}>` : mail.from;
  }
  return null;
};

export const parsedMailTo = (mail: Mail | IMAPEmail) => {
  if (mail) {
    const emailRegex = /\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b/;
    const match = mail.to?.toString().match(emailRegex);
    // @ts-ignore
    const name = mail.contact?.firstName || mail.data?.name;
    return match ? `${name?.replace(/[^\w\s]/g, "")} <${match[0]}>` : mail.to;
  }
  return null;
};
