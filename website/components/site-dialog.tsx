"use client";
import { useEffect, useRef } from "react";
import { X } from "lucide-react";
export function SiteDialog({
  title,
  onClose,
  children,
}: {
  title: string;
  onClose: () => void;
  children: React.ReactNode;
}) {
  const ref = useRef<HTMLDialogElement>(null);
  useEffect(() => {
    const d = ref.current;
    const active = document.activeElement as HTMLElement | null;
    const wasLocked = document.body.classList.contains("overflow-hidden");
    document.body.classList.add("overflow-hidden");
    d?.showModal();
    return () => {
      d?.close();
      if (!wasLocked) document.body.classList.remove("overflow-hidden");
      active?.focus();
    };
  }, []);
  return (
    <dialog
      ref={ref}
      aria-labelledby="dialog-title"
      onCancel={onClose}
      onKeyDown={(event) => {
        if (event.key !== "Tab") return;
        const focusable = Array.from(
          event.currentTarget.querySelectorAll<HTMLElement>(
            'button:not([disabled]), a[href], input:not([disabled]), [tabindex]:not([tabindex="-1"])',
          ),
        ).filter((element) => element.getClientRects().length > 0);
        const first = focusable[0];
        const last = focusable[focusable.length - 1];
        if (event.shiftKey && document.activeElement === first) {
          event.preventDefault();
          last?.focus();
        } else if (!event.shiftKey && document.activeElement === last) {
          event.preventDefault();
          first?.focus();
        }
      }}
      onClick={(e) => {
        if (e.target === e.currentTarget) onClose();
      }}
      className="m-auto max-h-[92dvh] w-[min(1120px,94vw)] overflow-y-auto overscroll-contain rounded-2xl bg-cream p-0 text-ink shadow-2xl backdrop:bg-ink/65 backdrop:backdrop-blur-sm"
    >
      <div className="sticky top-0 z-10 flex items-center justify-between gap-4 border-b border-ink/10 bg-cream px-5 py-4 md:px-7">
        <h2 id="dialog-title" className="text-lg font-medium">
          {title}
        </h2>
        <button
          autoFocus
          onClick={onClose}
          aria-label="Close preview"
          className="rounded-full p-2 transition-colors hover:bg-ink/10"
        >
          <X size={22} />
        </button>
      </div>
      {children}
    </dialog>
  );
}
