"use client";

// Compatibility names for existing feature imports. Every overlay uses the
// Coss inset sheet foundation; there is no separate centered-dialog surface.
export {
  Sheet as Dialog,
  SheetPortal as DialogPortal,
  SheetOverlay as DialogOverlay,
  SheetClose as DialogClose,
  SheetTrigger as DialogTrigger,
  SheetContent as DialogContent,
  SheetHeader as DialogHeader,
  SheetFooter as DialogFooter,
  SheetTitle as DialogTitle,
  SheetDescription as DialogDescription,
} from "@/components/ui/sheet";
