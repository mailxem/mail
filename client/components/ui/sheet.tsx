"use client";

// Adapted from coss ui's Sheet (MIT): https://coss.com/ui/r/sheet.json.
// Base UI owns focus trapping, dismissal, scroll locking and focus restoration.
// Utilities are translated to this project's Tailwind 3; inset is our default.
import * as React from "react";
import { Dialog as SheetPrimitive } from "@base-ui/react/dialog";
import { ScrollArea as ScrollAreaPrimitive } from "@base-ui/react/scroll-area";
import { mergeProps } from "@base-ui/react/merge-props";
import { useRender } from "@base-ui/react/use-render";
import { X } from "lucide-react";
import { cn } from "@/lib/utils";

const Sheet = SheetPrimitive.Root;
const SheetPortal = SheetPrimitive.Portal;
const LegacyContentContext = React.createContext(false);

type SheetTriggerProps = SheetPrimitive.Trigger.Props & { asChild?: boolean };
const SheetTrigger = React.forwardRef<HTMLButtonElement, SheetTriggerProps>(
  ({ asChild, children, render, onClick, ...props }, ref) => (
    <SheetPrimitive.Trigger
      data-slot="sheet-trigger"
      ref={ref}
      render={asChild ? React.Children.only(children) as React.ReactElement : render}
      {...props}
      onClick={(event) => {
        onClick?.(event);
        if (event.defaultPrevented) event.preventBaseUIHandler();
      }}
    >
      {asChild ? undefined : children}
    </SheetPrimitive.Trigger>
  ),
);
SheetTrigger.displayName = "SheetTrigger";

type SheetCloseProps = SheetPrimitive.Close.Props & { asChild?: boolean };
const SheetClose = React.forwardRef<HTMLButtonElement, SheetCloseProps>(
  ({ asChild, children, render, onClick, ...props }, ref) => (
    <SheetPrimitive.Close
      data-slot="sheet-close"
      ref={ref}
      render={asChild ? React.Children.only(children) as React.ReactElement : render}
      {...props}
      onClick={(event) => {
        onClick?.(event);
        if (event.defaultPrevented) event.preventBaseUIHandler();
      }}
    >
      {asChild ? undefined : children}
    </SheetPrimitive.Close>
  ),
);
SheetClose.displayName = "SheetClose";

const SheetBackdrop = React.forwardRef<HTMLDivElement, SheetPrimitive.Backdrop.Props>(
  ({ className, ...props }, ref) => (
    <SheetPrimitive.Backdrop
      {...props}
      ref={ref}
      data-slot="sheet-backdrop"
      className={(state) => cn(
        "fixed inset-0 z-50 bg-black/40 backdrop-blur-[2px] transition-opacity duration-200 data-[ending-style]:opacity-0 data-[starting-style]:opacity-0 motion-reduce:transition-none",
        typeof className === "function" ? className(state) : className,
      )}
    />
  ),
);
SheetBackdrop.displayName = "SheetBackdrop";

type SheetSide = "right" | "left" | "top" | "bottom";
type SheetVariant = "default" | "inset";
function SheetViewport({
  className, side = "right", variant = "inset", ...props
}: SheetPrimitive.Viewport.Props & { side?: SheetSide; variant?: SheetVariant }) {
  return (
    <SheetPrimitive.Viewport
      {...props}
      data-slot="sheet-viewport"
      className={(state) => cn(
        "fixed inset-0 z-50 flex min-h-0 min-w-0",
        side === "right" && "justify-end",
        side === "left" && "justify-start",
        side === "top" && "items-start",
        side === "bottom" && "items-end",
        variant === "inset" && "p-2 sm:p-3",
        typeof className === "function" ? className(state) : className,
      )}
    />
  );
}

type SheetPopupProps = SheetPrimitive.Popup.Props & {
  side?: SheetSide;
  variant?: SheetVariant;
  showCloseButton?: boolean;
  closeProps?: SheetCloseProps;
  portalProps?: SheetPrimitive.Portal.Props;
  /** Compatibility with existing Radix consumers. */
  onOpenAutoFocus?: (event: Event) => void;
  onCloseAutoFocus?: (event: Event) => void;
};

const SheetPopup = React.forwardRef<HTMLDivElement, SheetPopupProps>(
  ({
    className, children, side = "right", variant = "inset", showCloseButton = true,
    closeProps, portalProps, initialFocus, finalFocus, onOpenAutoFocus,
    onCloseAutoFocus, style, ...props
  }, ref) => {
    const menuTrigger = React.useRef<HTMLElement | null>(null);
    const popupRef = React.useCallback((node: HTMLDivElement | null) => {
      if (node) {
        const active = document.activeElement instanceof HTMLElement ? document.activeElement : null;
        const triggerId = active?.closest('[role="menu"]')?.getAttribute("aria-labelledby");
        menuTrigger.current = triggerId ? document.getElementById(triggerId) : null;
      }
      if (typeof ref === "function") ref(node);
      else if (ref) ref.current = node;
    }, [ref]);
    const focusCompatibility = (handler: (event: Event) => void) => () => {
      const event = new Event("sheet.autofocus", { cancelable: true });
      handler(event);
      return !event.defaultPrevented;
    };
    return (
      <SheetPortal {...portalProps}>
        <SheetBackdrop />
        <SheetViewport side={side} variant={variant}>
          <SheetPrimitive.Popup
            {...props}
            ref={popupRef}
            data-slot="sheet-popup"
            data-side={side}
            data-variant={variant}
            initialFocus={onOpenAutoFocus ? focusCompatibility(onOpenAutoFocus) : initialFocus}
            finalFocus={onCloseAutoFocus ? focusCompatibility(onCloseAutoFocus) : finalFocus ?? (() =>
              menuTrigger.current?.isConnected ? menuTrigger.current : true
            )}
            className={(state) => cn(
              "relative flex min-h-0 min-w-0 w-full flex-col overflow-hidden border border-border bg-popover text-popover-foreground shadow-2xl outline-none transition-[opacity,transform] duration-200 ease-out data-[ending-style]:opacity-0 data-[starting-style]:opacity-0 motion-reduce:transition-none",
              (side === "left" || side === "right") && "h-full max-w-xl",
              (side === "top" || side === "bottom") && "max-h-[calc(100dvh-4rem)]",
              side === "right" && "data-[ending-style]:translate-x-8 data-[starting-style]:translate-x-8",
              side === "left" && "data-[ending-style]:-translate-x-8 data-[starting-style]:-translate-x-8",
              side === "top" && "data-[ending-style]:-translate-y-8 data-[starting-style]:-translate-y-8",
              side === "bottom" && "data-[ending-style]:translate-y-8 data-[starting-style]:translate-y-8",
              variant === "inset" && "rounded-2xl",
              typeof className === "function" ? className(state) : className,
            )}
            style={(state) => ({
              ...(typeof style === "function" ? style(state) : style),
              // Legacy h-screen/min-w utilities must not escape the inset viewport.
              minWidth: 0,
              maxHeight: "100%",
              ...(side === "left" || side === "right" ? { height: "100%" } : {}),
            })}
          >
            <LegacyContentContext.Provider value={false}>{children}</LegacyContentContext.Provider>
            {showCloseButton && (
              <SheetClose
                aria-label="Close sheet"
                {...closeProps}
                className={cn(
                  "absolute right-3 top-3 z-10 inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-lg text-muted-foreground transition-colors hover:bg-accent hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:pointer-events-none motion-reduce:transition-none",
                  typeof closeProps?.className === "string" ? closeProps.className : undefined,
                )}
              >
                <X className="h-4 w-4" aria-hidden="true" />
              </SheetClose>
            )}
          </SheetPrimitive.Popup>
        </SheetViewport>
      </SheetPortal>
    );
  },
);
SheetPopup.displayName = "SheetPopup";

// Legacy content includes its own forms/layouts instead of Coss SheetPanel.
// Keep it scrollable and padded while exposing the new explicit composition.
const SheetContent = React.forwardRef<HTMLDivElement, SheetPopupProps>(
  ({ className, children, ...props }, ref) => (
    <SheetPopup
      {...props}
      ref={ref}
      className={(state) => cn(
        "gap-6 overflow-y-auto overscroll-contain p-6 [&>[data-slot=sheet-footer]]:mt-auto",
        typeof className === "function" ? className(state) : className,
      )}
    >
      <LegacyContentContext.Provider value={true}>{children}</LegacyContentContext.Provider>
    </SheetPopup>
  ),
);
SheetContent.displayName = "SheetContent";

function SheetHeader({ className, render, ...props }: useRender.ComponentProps<"div">) {
  const legacy = React.useContext(LegacyContentContext);
  const defaults = {
    className: cn("flex shrink-0 flex-col gap-2 text-left", legacy ? "p-0 pr-8" : "p-6 pr-14", className),
    "data-slot": "sheet-header",
  };
  return useRender({ defaultTagName: "div", render, props: mergeProps<"div">(defaults, props) });
}

function SheetFooter({ className, variant = "default", render, ...props }:
  useRender.ComponentProps<"div"> & { variant?: "default" | "bare" }) {
  const defaults = {
    className: cn(
      "flex shrink-0 flex-col-reverse gap-2 px-6 py-4 sm:flex-row sm:justify-end",
      variant === "default" && "border-t border-border bg-muted/50",
      className,
    ),
    "data-slot": "sheet-footer",
  };
  return useRender({ defaultTagName: "div", render, props: mergeProps<"div">(defaults, props) });
}

function SheetPanel({ className, scrollFade = true, render, ...props }:
  useRender.ComponentProps<"div"> & { scrollFade?: boolean }) {
  const defaults = {
    className: cn("px-6 pb-6 pt-1", className),
    "data-slot": "sheet-panel",
  };
  const content = useRender({ defaultTagName: "div", render, props: mergeProps<"div">(defaults, props) });
  return (
    <ScrollAreaPrimitive.Root className="relative min-h-0 flex-1" overflowEdgeThreshold={8}>
      <ScrollAreaPrimitive.Viewport
        className="h-full overscroll-contain outline-none focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-ring"
        style={(state) => ({
          maskImage: scrollFade && (state.overflowYStart || state.overflowYEnd)
            ? `linear-gradient(to bottom, ${state.overflowYStart ? "transparent, black 12px" : "black 0"}, ${state.overflowYEnd ? "black calc(100% - 12px), transparent" : "black 100%"})`
            : undefined,
        })}
      >
        <ScrollAreaPrimitive.Content>{content}</ScrollAreaPrimitive.Content>
      </ScrollAreaPrimitive.Viewport>
      <ScrollAreaPrimitive.Scrollbar className="m-1 flex w-1.5 touch-none select-none rounded-full opacity-0 transition-opacity data-[hovering]:opacity-100 data-[scrolling]:opacity-100 motion-reduce:transition-none">
        <ScrollAreaPrimitive.Thumb className="w-full rounded-full bg-border" />
      </ScrollAreaPrimitive.Scrollbar>
    </ScrollAreaPrimitive.Root>
  );
}

const SheetTitle = React.forwardRef<HTMLHeadingElement, SheetPrimitive.Title.Props>(
  ({ className, ...props }, ref) => (
    <SheetPrimitive.Title
      {...props}
      ref={ref}
      data-slot="sheet-title"
      className={(state) => cn("text-xl font-medium leading-tight tracking-[-0.025em]", typeof className === "function" ? className(state) : className)}
    />
  ),
);
SheetTitle.displayName = "SheetTitle";

const SheetDescription = React.forwardRef<HTMLParagraphElement, SheetPrimitive.Description.Props>(
  ({ className, ...props }, ref) => (
    <SheetPrimitive.Description
      {...props}
      ref={ref}
      data-slot="sheet-description"
      className={(state) => cn("text-sm leading-relaxed text-muted-foreground", typeof className === "function" ? className(state) : className)}
    />
  ),
);
SheetDescription.displayName = "SheetDescription";

export {
  Sheet, SheetPrimitive, SheetPortal, SheetBackdrop, SheetBackdrop as SheetOverlay,
  SheetViewport, SheetTrigger, SheetClose, SheetPopup, SheetContent, SheetHeader,
  SheetFooter, SheetPanel, SheetTitle, SheetDescription,
};
export type { SheetPopupProps, SheetCloseProps };
