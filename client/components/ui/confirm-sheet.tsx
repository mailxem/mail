"use client";

import * as React from "react";
import {
  AlertDialog, AlertDialogContent, AlertDialogHeader, AlertDialogTitle,
  AlertDialogDescription, AlertDialogFooter, AlertDialogCancel, AlertDialogAction,
} from "@/components/ui/alert-dialog";
import { buttonVariants } from "@/components/ui/button";

export type ConfirmSheetOptions = {
  title?: string;
  description: string;
  confirmLabel?: string;
  cancelLabel?: string;
  variant?: "default" | "destructive";
};
type ConfirmSheet = (options: string | ConfirmSheetOptions) => Promise<boolean>;
type Request = {
  id: number;
  options: ConfirmSheetOptions;
  resolve: (value: boolean) => void;
  returnFocus: HTMLElement | null;
};

function getReturnFocus(): HTMLElement | null {
  const active = document.activeElement instanceof HTMLElement ? document.activeElement : null;
  // Menu items unmount as soon as a confirmation opens. Restore to the menu's
  // trigger instead; otherwise the browser drops keyboard users on <body>.
  const menu = active?.closest('[role="menu"]');
  const triggerId = menu?.getAttribute("aria-labelledby");
  return (triggerId ? document.getElementById(triggerId) : null) ?? active;
}
const ConfirmSheetContext = React.createContext<ConfirmSheet | null>(null);

/** Await a user's decision before running the existing API operation. */
export function useConfirmSheet(): ConfirmSheet {
  const confirm = React.useContext(ConfirmSheetContext);
  if (!confirm) throw new Error("useConfirmSheet must be used inside ConfirmSheetProvider");
  return confirm;
}

export function ConfirmSheetProvider({ children }: { children: React.ReactNode }) {
  const pending = React.useRef<Request[]>([]);
  const sequence = React.useRef(0);
  const [requests, setRequests] = React.useState<Request[]>([]);
  const confirm = React.useCallback<ConfirmSheet>((input) => new Promise((resolve) => {
    const options = typeof input === "string" ? { description: input } : input;
    const returnFocus = pending.current[0]?.returnFocus ?? getReturnFocus();
    pending.current = [...pending.current, { id: ++sequence.current, options, resolve, returnFocus }];
    setRequests(pending.current);
  }), []);
  const settle = React.useCallback((id: number, result: boolean) => {
    const request = pending.current.find((item) => item.id === id);
    if (!request) return;
    pending.current = pending.current.filter((item) => item.id !== id);
    setRequests(pending.current);
    request.resolve(result);
  }, []);
  React.useEffect(() => () => {
    pending.current.forEach((request) => request.resolve(false));
    pending.current = [];
  }, []);
  const active = requests[0];
  const lastRequest = React.useRef<Request | null>(null);
  React.useEffect(() => { if (active) lastRequest.current = active; }, [active]);
  // Keep the decided request visible through the closing transition.
  const current = active ?? lastRequest.current;
  return (
    <ConfirmSheetContext.Provider value={confirm}>
      {children}
      <AlertDialog
        open={!!active}
        onOpenChange={(open) => { if (!open && active) settle(active.id, false); }}
      >
        <AlertDialogContent finalFocus={() => current?.returnFocus?.isConnected ? current.returnFocus : true}>
          <AlertDialogHeader>
            <AlertDialogTitle>{current?.options.title ?? "Confirm this action"}</AlertDialogTitle>
            <AlertDialogDescription>{current?.options.description}</AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel onClick={() => current && settle(current.id, false)}>
              {current?.options.cancelLabel ?? "Cancel"}
            </AlertDialogCancel>
            <AlertDialogAction
              className={buttonVariants({ variant: current?.options.variant ?? "destructive" })}
              onClick={() => current && settle(current.id, true)}
            >
              {current?.options.confirmLabel ?? "Confirm"}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </ConfirmSheetContext.Provider>
  );
}
