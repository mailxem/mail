/* Xem forms: one script supports inline forms and inset sheets (including legacy aliases). */
(function () {
  "use strict";
  var script = document.currentScript;
  if (!script || !script.src) return;
  var allowedOrigin;
  try {
    allowedOrigin = new URL(script.src).origin;
  } catch (_) {
    return;
  }
  // A duplicate script include must not add duplicate launchers or message handlers.
  var marker = "data-xem-initialized";

  function mount(container) {
    if (container.hasAttribute(marker)) return;
    var url;
    try {
      url = new URL(container.getAttribute("data-xem-form"), allowedOrigin);
      if (
        url.origin !== allowedOrigin ||
        !/^\/f\/[^/]+\/?$/.test(url.pathname) ||
        !/^https?:$/.test(url.protocol) ||
        url.username ||
        url.password
      )
        return;
    } catch (_) {
      return;
    }
    container.setAttribute(marker, "true");
    var mode = container.getAttribute("data-xem-mode") || "inline";
    if (["inline", "sheet", "popup", "sidebar"].indexOf(mode) === -1) mode = "inline";
    var label = container.getAttribute("data-xem-label") || "Open form";
    // Forward campaign attribution only, never arbitrary page parameters or sensitive answers.
    var campaign = new URLSearchParams(window.location.search);
    [
      "utm_source",
      "utm_medium",
      "utm_campaign",
      "utm_content",
      "utm_term",
    ].forEach(function (key) {
      if (campaign.has(key))
        url.searchParams.set(key, campaign.get(key).slice(0, 200));
    });
    url.searchParams.set("embedOrigin", window.location.origin);
    var frame = document.createElement("iframe");
    frame.title = container.getAttribute("data-xem-title") || "Xem form";
    frame.loading = "lazy";
    frame.referrerPolicy = "strict-origin";
    frame.setAttribute(
      "sandbox",
      "allow-scripts allow-forms allow-same-origin allow-popups",
    );
    frame.style.cssText =
      "display:block;width:100%;height:640px;min-height:320px;border:0;background:transparent;";
    frame.src = url.href;
    var dialog = null;
    var launcher = null;
    var closeButton = null;
    var previousFocus = null;
    var originalOverflow = "";
    function close() {
      if (!dialog || !dialog.open) return;
      dialog.close();
      document.documentElement.style.overflow = originalOverflow;
      if (previousFocus && previousFocus.isConnected) previousFocus.focus();
      if (launcher) launcher.setAttribute("aria-expanded", "false");
    }
    function open() {
      if (!dialog || dialog.open) return;
      // Close this script's other open forms to avoid nested focus traps.
      document
        .querySelectorAll("dialog[data-xem-dialog][open]")
        .forEach(function (other) {
          other.dispatchEvent(new Event("xem:close"));
        });
      previousFocus = document.activeElement;
      originalOverflow = document.documentElement.style.overflow;
      document.documentElement.style.overflow = "hidden";
      dialog.showModal();
      if (launcher) launcher.setAttribute("aria-expanded", "true");
      closeButton.focus();
    }
    if (mode === "inline") {
      container.appendChild(frame);
    } else {
      launcher = document.createElement("button");
      launcher.type = "button";
      launcher.textContent = label;
      launcher.setAttribute("aria-haspopup", "dialog");
      launcher.setAttribute("aria-expanded", "false");
      launcher.style.cssText =
        "font:inherit;cursor:pointer;padding:12px 20px;border:1px solid currentColor;border-radius:8px;";
      dialog = document.createElement("dialog");
      dialog.setAttribute("data-xem-dialog", "true");
      dialog.setAttribute("aria-label", frame.title);
      dialog.style.cssText =
        "position:fixed;inset:12px 12px 12px auto;box-sizing:border-box;padding:0;border:1px solid #ddd;background:#fff;color:#222;width:min(560px,calc(100vw - 24px));height:calc(100dvh - 24px);max-width:calc(100vw - 24px);max-height:calc(100dvh - 24px);margin:0;border-radius:16px;overflow:auto;overscroll-behavior:contain;";
      closeButton = document.createElement("button");
      closeButton.type = "button";
      closeButton.textContent = "Close";
      closeButton.setAttribute("aria-label", "Close form");
      closeButton.style.cssText =
        "display:block;position:sticky;top:0;margin:8px 12px 8px auto;padding:10px 14px;min-height:44px;background:#fff;color:#222;border:1px solid #bbb;border-radius:6px;cursor:pointer;font:14px system-ui;z-index:1;";
      closeButton.addEventListener("click", close);
      launcher.addEventListener("click", open);
      dialog.addEventListener("cancel", function (event) {
        event.preventDefault();
        close();
      });
      dialog.addEventListener("xem:close", close);
      dialog.addEventListener("click", function (event) {
        if (event.target !== dialog) return;
        var bounds = dialog.getBoundingClientRect();
        if (
          event.clientX < bounds.left ||
          event.clientX > bounds.right ||
          event.clientY < bounds.top ||
          event.clientY > bounds.bottom
        )
          close();
      });
      dialog.appendChild(closeButton);
      dialog.appendChild(frame);
      container.appendChild(launcher);
      document.body.appendChild(dialog);
    }
    window.addEventListener("message", function (event) {
      if (
        event.origin !== allowedOrigin ||
        event.source !== frame.contentWindow ||
        !event.data ||
        typeof event.data !== "object"
      )
        return;
      if (
        event.data.type === "xem:resize" &&
        typeof event.data.height === "number" &&
        Number.isFinite(event.data.height)
      ) {
        frame.style.height =
          Math.min(10000, Math.max(320, event.data.height)) + "px";
      }
      if (event.data.type === "xem:complete")
        container.dispatchEvent(
          new CustomEvent("xem:complete", { bubbles: true }),
        );
    });
  }
  function initialize() {
    document.querySelectorAll("[data-xem-form]").forEach(mount);
  }
  if (document.readyState === "loading")
    document.addEventListener("DOMContentLoaded", initialize, { once: true });
  else initialize();
})();
