"use client";

import * as React from "react";
import { AlertDialog as AlertPrimitive } from "@base-ui/react/alert-dialog";
import {
  SheetPortal, SheetOverlay, SheetTrigger, SheetContent, SheetHeader,
  SheetFooter, SheetTitle, SheetDescription, SheetClose,
  type SheetPopupProps, type SheetCloseProps,
} from "@/components/ui/sheet";
import { buttonVariants } from "@/components/ui/button";
import { cn } from "@/lib/utils";

// Base UI's alert root keeps destructive confirmations modal and rejects
// backdrop dismissal. Its popup is the same accessible Coss inset sheet.
const AlertDialog = AlertPrimitive.Root;
const AlertDialogPortal = SheetPortal;
const AlertDialogOverlay = SheetOverlay;
const AlertDialogTrigger = SheetTrigger;
const AlertDialogHeader = SheetHeader;
const AlertDialogFooter = SheetFooter;
const AlertDialogTitle = SheetTitle;
const AlertDialogDescription = SheetDescription;

const AlertDialogContent = React.forwardRef<HTMLDivElement, SheetPopupProps>(
  ({ initialFocus, ...props }, forwardedRef) => {
    const popupRef = React.useRef<HTMLDivElement>(null);
    React.useImperativeHandle(forwardedRef, () => popupRef.current!);
    return (
      <SheetContent
        {...props}
        ref={popupRef}
        showCloseButton={false}
        initialFocus={initialFocus ?? (() =>
          popupRef.current?.querySelector<HTMLElement>('[data-slot="alert-sheet-cancel"]') ?? true
        )}
      />
    );
  },
);
AlertDialogContent.displayName = "AlertDialogContent";

const AlertDialogAction = React.forwardRef<HTMLButtonElement, SheetCloseProps>(
  ({ className, ...props }, ref) => (
    <SheetClose
      {...props}
      ref={ref}
      className={(state) => cn(buttonVariants(), typeof className === "function" ? className(state) : className)}
    />
  ),
);
AlertDialogAction.displayName = "AlertDialogAction";

const AlertDialogCancel = React.forwardRef<HTMLButtonElement, SheetCloseProps>(
  ({ className, ...props }, ref) => (
    <SheetClose
      {...props}
      ref={ref}
      data-slot="alert-sheet-cancel"
      className={(state) => cn(buttonVariants({ variant: "outline" }), typeof className === "function" ? className(state) : className)}
    />
  ),
);
AlertDialogCancel.displayName = "AlertDialogCancel";

export {
  AlertDialog, AlertDialogPortal, AlertDialogOverlay, AlertDialogTrigger,
  AlertDialogContent, AlertDialogHeader, AlertDialogFooter, AlertDialogTitle,
  AlertDialogDescription, AlertDialogAction, AlertDialogCancel,
};
