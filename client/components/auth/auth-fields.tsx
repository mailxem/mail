"use client";
import { forwardRef, useState, type ComponentProps } from "react";
import { Eye, EyeOff } from "lucide-react";
import styles from "./auth.module.css";

type Props = ComponentProps<"input"> & { label: string; error?: string };
export const AuthField = forwardRef<HTMLInputElement, Props>(function AuthField(
  { label, error, id, ...props },
  ref,
) {
  const inputID = id || props.name;
  return (
    <div className={styles.field}>
      <label htmlFor={inputID}>{label}</label>
      <input
        {...props}
        id={inputID}
        ref={ref}
        aria-invalid={!!error}
        aria-describedby={error ? `${inputID}-error` : undefined}
      />
      {error && (
        <p className={styles.fieldError} id={`${inputID}-error`}>
          {error}
        </p>
      )}
    </div>
  );
});
export const AuthPasswordField = forwardRef<HTMLInputElement, Props>(
  function AuthPasswordField({ label, error, id, ...props }, ref) {
    const [visible, setVisible] = useState(false);
    const inputID = id || props.name;
    return (
      <div className={styles.field}>
        <label htmlFor={inputID}>{label}</label>
        <div className={styles.password}>
          <input
            {...props}
            id={inputID}
            ref={ref}
            type={visible ? "text" : "password"}
            aria-invalid={!!error}
            aria-describedby={error ? `${inputID}-error` : undefined}
          />
          <button
            className={styles.reveal}
            type="button"
            aria-label={`${visible ? "Hide" : "Show"} ${label.toLowerCase()}`}
            aria-pressed={visible}
            onClick={() => setVisible(!visible)}
          >
            {visible ? <EyeOff size={15} /> : <Eye size={15} />}
          </button>
        </div>
        {error && (
          <p className={styles.fieldError} id={`${inputID}-error`}>
            {error}
          </p>
        )}
      </div>
    );
  },
);
