"use client";

import { useId, useRef, useState } from "react";
import { Input, type InputProps } from "@/components/ui/input";
import { fieldQuery, insertField, matchFields } from "@/lib/email/fields";
import { cn } from "@/lib/utils";

/** A text input where typing {{ suggests contact fields, narrowed as you type. */
export function FieldSuggestInput({
  value,
  onValueChange,
  onBlur,
  ...props
}: Omit<InputProps, "value" | "onChange"> & {
  value: string;
  onValueChange: (value: string) => void;
}) {
  const ref = useRef<HTMLInputElement>(null);
  const listId = useId();
  const [query, setQuery] = useState<string>();
  const [active, setActive] = useState(0);
  const options = query === undefined ? [] : matchFields(query);

  const sync = (el: HTMLInputElement) => {
    const next = fieldQuery(el.value, el.selectionStart ?? el.value.length);
    if (next !== query) setActive(0);
    setQuery(next);
  };
  const pick = (key: string) => {
    const el = ref.current;
    if (!el) return;
    const next = insertField(value, el.selectionStart ?? value.length, key);
    onValueChange(next.text);
    setQuery(undefined);
    requestAnimationFrame(() => {
      el.focus();
      el.setSelectionRange(next.caret, next.caret);
    });
  };

  return (
    <div className="relative">
      <Input
        {...props}
        ref={ref}
        value={value}
        role="combobox"
        aria-autocomplete="list"
        aria-expanded={options.length > 0}
        aria-controls={listId}
        aria-activedescendant={
          options.length ? `${listId}-${options[active].key}` : undefined
        }
        onChange={(e) => {
          onValueChange(e.target.value);
          sync(e.target);
        }}
        onSelect={(e) => sync(e.currentTarget)}
        onBlur={(e) => {
          setQuery(undefined);
          onBlur?.(e);
        }}
        onKeyDown={(e) => {
          if (!options.length) return;
          if (e.key === "ArrowDown" || e.key === "ArrowUp") {
            e.preventDefault();
            const step = e.key === "ArrowDown" ? 1 : -1;
            setActive((active + step + options.length) % options.length);
          } else if (e.key === "Enter" || e.key === "Tab") {
            e.preventDefault();
            pick(options[active].key);
          } else if (e.key === "Escape") {
            setQuery(undefined);
          }
        }}
      />
      {options.length > 0 && (
        <div className="absolute z-50 mt-1 w-64 rounded-lg border border-border bg-popover p-1 text-sm shadow-md">
          <ul id={listId} role="listbox" aria-label="Contact fields">
            {options.map((field, i) => (
              <li
                key={field.key}
                id={`${listId}-${field.key}`}
                role="option"
                aria-selected={i === active}
                onMouseDown={(e) => {
                  e.preventDefault();
                  pick(field.key);
                }}
                className={cn(
                  "flex cursor-pointer items-center justify-between gap-3 rounded-md px-2 py-1.5",
                  i === active && "bg-muted",
                )}
              >
                <span>{field.label}</span>
                <span className="font-mono text-xs text-muted-foreground">
                  {field.key}
                </span>
              </li>
            ))}
          </ul>
          <p className="px-2 pb-1 pt-1.5 text-xs text-muted-foreground">
            Add a fallback for readers without one:{" "}
            <span className="font-mono">{"{{first_name|there}}"}</span>
          </p>
        </div>
      )}
    </div>
  );
}
