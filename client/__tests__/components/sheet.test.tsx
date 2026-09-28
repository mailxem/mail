/** @jest-environment jsdom */
import * as React from "react";
import { act } from "react";
import { createRoot, Root } from "react-dom/client";
import { ConfirmSheetProvider, useConfirmSheet } from "@/components/ui/confirm-sheet";
import { Dialog, DialogContent, DialogTitle, DialogTrigger, DialogClose } from "@/components/ui/dialog";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@/components/ui/dropdown-menu";

let root: Root;
let container: HTMLDivElement;
let confirm: ReturnType<typeof useConfirmSheet>;
function CaptureConfirm() { confirm = useConfirmSheet(); return null; }
async function render(content: React.ReactNode) {
  await act(async () => { root.render(content); });
}
async function click(element: Element) {
  await act(async () => { element.dispatchEvent(new MouseEvent("click", { bubbles: true })); });
}
async function flush() {
  await act(async () => { await new Promise((resolve) => setTimeout(resolve, 30)); });
}
function button(text: string) {
  const element = [...document.querySelectorAll("button")].find((item) => item.textContent === text);
  if (!element) throw new Error(`Missing button: ${text}`);
  return element;
}
beforeAll(() => {
  (globalThis as any).IS_REACT_ACT_ENVIRONMENT = true;
  globalThis.ResizeObserver = class { observe() {} unobserve() {} disconnect() {} };
  Object.defineProperty(HTMLElement.prototype, "getClientRects", {
    configurable: true,
    value: () => [{ width: 100, height: 30, top: 0, left: 0, bottom: 30, right: 100 }],
  });
  HTMLElement.prototype.scrollIntoView = () => {};
});
beforeEach(() => {
  container = document.createElement("div");
  document.body.appendChild(container);
  root = createRoot(container);
});
afterEach(async () => {
  await act(async () => { root.unmount(); });
  container.remove();
});

test("legacy asChild trigger renders one button and close autofocus remains cancelable", async () => {
  const restore = document.createElement("button");
  document.body.appendChild(restore);
  await render(
    <Dialog>
      <DialogTrigger asChild><button>Open details</button></DialogTrigger>
      <DialogContent onCloseAutoFocus={(event) => { event.preventDefault(); restore.focus(); }}>
        <DialogTitle>Details</DialogTitle>
        <DialogClose>Done</DialogClose>
      </DialogContent>
    </Dialog>,
  );
  expect(document.querySelectorAll("button button")).toHaveLength(0);
  await click(button("Open details"));
  await flush();
  const popup = document.querySelector('[role="dialog"]');
  expect(popup?.getAttribute("data-variant")).toBe("inset");
  expect(popup?.getAttribute("data-side")).toBe("right");
  await click(button("Done"));
  await flush();
  expect(document.activeElement).toBe(restore);
  restore.remove();
});

test("preventDefault keeps an existing confirmation action open while a consumer is busy", async () => {
  await render(
    <Dialog defaultOpen>
      <DialogContent><DialogTitle>Save changes</DialogTitle>
        <DialogClose onClick={(event) => event.preventDefault()}>Save</DialogClose>
      </DialogContent>
    </Dialog>,
  );
  await click(button("Save"));
  expect(document.querySelector('[role="dialog"]')).not.toBeNull();
});

test("confirmation queue settles exactly one request and focuses Cancel for each decision", async () => {
  await render(<ConfirmSheetProvider><CaptureConfirm /></ConfirmSheetProvider>);
  let first: Promise<boolean>;
  let second: Promise<boolean>;
  await act(async () => {
    first = confirm({ title: "First decision", description: "First", confirmLabel: "Delete first" });
    second = confirm({ title: "Second decision", description: "Second", confirmLabel: "Delete second" });
  });
  await flush();
  expect(document.activeElement).toBe(button("Cancel"));
  await click(button("Delete first"));
  await expect(first!).resolves.toBe(true);
  expect(document.querySelector('[role="alertdialog"]')?.textContent).toContain("Second decision");
  expect(document.activeElement).toBe(button("Cancel"));
  await click(button("Cancel"));
  await expect(second!).resolves.toBe(false);
});

test("Escape cancels and provider unmount resolves every pending request", async () => {
  await render(<ConfirmSheetProvider><CaptureConfirm /></ConfirmSheetProvider>);
  let escaped: Promise<boolean>;
  await act(async () => { escaped = confirm("Do this?"); });
  await flush();
  await act(async () => {
    document.activeElement?.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape", bubbles: true }));
  });
  await expect(escaped!).resolves.toBe(false);
  let one: Promise<boolean>;
  let two: Promise<boolean>;
  await act(async () => { one = confirm("One"); two = confirm("Two"); });
  await render(null);
  await expect(one!).resolves.toBe(false);
  await expect(two!).resolves.toBe(false);
});

function MenuConfirmation() {
  const request = useConfirmSheet();
  return <DropdownMenu defaultOpen>
    <DropdownMenuTrigger>Options</DropdownMenuTrigger>
    <DropdownMenuContent>
      <DropdownMenuItem onClick={() => { void request("Delete this connection?"); }}>Delete connection</DropdownMenuItem>
    </DropdownMenuContent>
  </DropdownMenu>;
}
test("a closing dropdown hands focus to the confirmation sheet", async () => {
  await render(<ConfirmSheetProvider><MenuConfirmation /></ConfirmSheetProvider>);
  await flush();
  await click(document.querySelector('[role="menuitem"]')!);
  await flush();
  expect(document.querySelector('[role="menu"]')).toBeNull();
  expect(document.querySelector('[role="alertdialog"]')).not.toBeNull();
  expect(document.activeElement).toBe(button("Cancel"));
  await click(button("Cancel"));
  await flush();
  expect(document.activeElement).toBe(button("Options"));
});

async function outsidePress(element: Element) {
  await act(async () => {
    element.dispatchEvent(new MouseEvent("pointerdown", { bubbles: true, button: 0 }));
    element.dispatchEvent(new MouseEvent("mousedown", { bubbles: true, button: 0 }));
    element.dispatchEvent(new MouseEvent("pointerup", { bubbles: true, button: 0 }));
    element.dispatchEvent(new MouseEvent("mouseup", { bubbles: true, button: 0 }));
    element.dispatchEvent(new MouseEvent("click", { bubbles: true, button: 0 }));
  });
}
test("ordinary sheet dismisses from backdrop but destructive confirmation does not", async () => {
  await render(<Dialog defaultOpen><DialogContent><DialogTitle>Details</DialogTitle></DialogContent></Dialog>);
  await flush();
  await outsidePress(document.querySelector('[data-slot="sheet-backdrop"]')!);
  await flush();
  expect(document.querySelector('[role="dialog"]')).toBeNull();
  await render(<ConfirmSheetProvider><CaptureConfirm /></ConfirmSheetProvider>);
  let result: Promise<boolean>;
  await act(async () => { result = confirm("Really remove this?"); });
  await flush();
  await outsidePress(document.querySelector('[data-slot="sheet-backdrop"]')!);
  await flush();
  expect(document.querySelector('[role="alertdialog"]')).not.toBeNull();
  await click(button("Cancel"));
  await expect(result!).resolves.toBe(false);
});

test("controlled sheets let the consumer reject dismissal while saving", async () => {
  const onOpenChange = jest.fn();
  await render(<Dialog open onOpenChange={onOpenChange}><DialogContent><DialogTitle>Saving</DialogTitle></DialogContent></Dialog>);
  await flush();
  await click(document.querySelector('[aria-label="Close sheet"]')!);
  expect(onOpenChange).toHaveBeenCalledWith(false, expect.any(Object));
  expect(document.querySelector('[role="dialog"]')).not.toBeNull();
});

test("a Radix select inside a sheet remains selectable and returns focus to its trigger", async () => {
  const onValueChange = jest.fn();
  await render(<Dialog defaultOpen><DialogContent><DialogTitle>Configure</DialogTitle>
    <Select onValueChange={onValueChange}>
      <SelectTrigger aria-label="Provider"><SelectValue placeholder="Choose provider" /></SelectTrigger>
      <SelectContent><SelectItem value="smtp">SMTP</SelectItem><SelectItem value="ses">SES</SelectItem></SelectContent>
    </Select>
  </DialogContent></Dialog>);
  await flush();
  const trigger = document.querySelector('[role="combobox"]')!;
  await act(async () => { trigger.dispatchEvent(new KeyboardEvent("keydown", { key: "ArrowDown", bubbles: true })); });
  await flush();
  const option = [...document.querySelectorAll('[role="option"]')].find((element) => element.textContent === "SES")!;
  expect(option.closest('[aria-hidden="true"]')).toBeNull();
  await act(async () => {
    (option as HTMLElement).focus();
    option.dispatchEvent(new KeyboardEvent("keydown", { key: "Enter", bubbles: true }));
  });
  await flush();
  expect(onValueChange).toHaveBeenCalledWith("ses");
  expect(document.activeElement).toBe(trigger);
  expect(document.querySelector('[role="dialog"]')).not.toBeNull();
});

function MenuDetails() {
  const [open, setOpen] = React.useState(false);
  return <><DropdownMenu defaultOpen>
    <DropdownMenuTrigger>Connection options</DropdownMenuTrigger>
    <DropdownMenuContent>
      <DropdownMenuItem onClick={() => setOpen(true)}>Edit connection</DropdownMenuItem>
    </DropdownMenuContent>
  </DropdownMenu><Dialog open={open} onOpenChange={setOpen}>
    <DialogContent><DialogTitle>Edit connection</DialogTitle><DialogClose>Done editing</DialogClose></DialogContent>
  </Dialog></>;
}
test("legacy edit sheets restore focus to the dropdown trigger after its item unmounts", async () => {
  await render(<MenuDetails />);
  await flush();
  await click(document.querySelector('[role="menuitem"]')!);
  await flush();
  await click(button("Done editing"));
  await flush();
  expect(document.activeElement).toBe(button("Connection options"));
});
