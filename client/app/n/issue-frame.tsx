"use client";

import { useEffect, useRef, useState } from "react";
import styles from "./archive.module.css";

const POLICY = `<meta http-equiv="Content-Security-Policy" content="default-src 'none'; img-src https: data:; style-src 'unsafe-inline'; font-src 'none'; base-uri 'none'; form-action 'none'"><base target="_blank">`;

/**
 * The email in an iframe with scripts disabled, so its HTML can't touch this
 * page. Same origin only so the frame can grow to the email's height.
 */
export function IssueFrame({ html, title }: { html: string; title: string }) {
  const ref = useRef<HTMLIFrameElement>(null);
  const [height, setHeight] = useState(600);
  const fit = () => {
    const body = ref.current?.contentDocument?.body;
    if (!body) return;
    // scrollHeight never drops below the frame's own height, so measure content.
    const bottom = Math.max(
      0,
      ...Array.from(body.children, (el) => el.getBoundingClientRect().bottom),
    );
    if (bottom > 0) setHeight(Math.ceil(bottom) + 16);
  };
  // The frame can finish loading before hydration attaches onLoad.
  useEffect(() => {
    if (ref.current?.contentDocument?.readyState === "complete") fit();
  }, [html]);
  return (
    <iframe
      ref={ref}
      title={title}
      sandbox="allow-same-origin allow-popups allow-popups-to-escape-sandbox"
      referrerPolicy="no-referrer"
      srcDoc={POLICY + html}
      className={styles.frame}
      style={{ height }}
      onLoad={fit}
    />
  );
}
