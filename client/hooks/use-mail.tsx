import { create } from "zustand";
import { Mail } from "@/app/types";
import { IMAPEmail } from "@/types/imap";

type MailState = {
  selected: string | null;
  mail: Mail | IMAPEmail | null;
  setMail: (state: {
    selected: string | null;
    mail: Mail | IMAPEmail | null;
  }) => void;
};

export const useMail = create<MailState>((set) => ({
  selected: null,
  mail: null,
  setMail: (newState) =>
    set(() => ({
      selected: newState.selected,
      mail: newState.mail,
    })),
}));
