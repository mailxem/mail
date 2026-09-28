import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { runInNewContext } from "node:vm";

class Element {
  attrs: Record<string, string> = {};
  children: Element[] = [];
  style: Record<string, string> = {};
  listeners: Record<string, Function[]> = {};
  contentWindow = {};
  title = "";
  src = "";
  open = false;
  textContent = "";
  isConnected = true;
  constructor(public tag: string) {}
  setAttribute(k: string, v: string) {
    this.attrs[k] = v;
  }
  getAttribute(k: string) {
    return this.attrs[k] ?? null;
  }
  hasAttribute(k: string) {
    return k in this.attrs;
  }
  appendChild(e: Element) {
    this.children.push(e);
  }
  addEventListener(k: string, fn: Function) {
    (this.listeners[k] ||= []).push(fn);
  }
  dispatchEvent(event: { type: string }) {
    for (const fn of this.listeners[event.type] || []) fn(event);
  }
  focus() {}
  showModal() {
    this.open = true;
  }
  close() {
    this.open = false;
  }
}

function fixture(urls: string[]) {
  const containers = urls.map((url) => {
    const element = new Element("div");
    element.setAttribute("data-xem-form", url);
    return element;
  });
  const listeners: Function[] = [];
  const body = new Element("body");
  const window = {
    location: {
      origin: "https://publisher.example",
      search: "?utm_source=newsletter&email=private@example.com",
    },
    addEventListener: (_: string, fn: Function) => listeners.push(fn),
  };
  const document = {
    currentScript: { src: "https://forms.example/forms/embed.js" },
    readyState: "complete",
    body,
    documentElement: new Element("html"),
    activeElement: null,
    createElement: (tag: string) => new Element(tag),
    querySelectorAll: (query: string) =>
      query === "[data-xem-form]"
        ? containers
        : body.children.filter((c) => c.open),
  };
  const context = {
    document,
    window,
    URL,
    URLSearchParams,
    Number,
    Event: class {
      constructor(public type: string) {}
    },
    CustomEvent: class {
      constructor(
        public type: string,
        public options: object,
      ) {}
    },
  };
  const run = () =>
    runInNewContext(
      readFileSync(resolve(process.cwd(), "public/forms/embed.js"), "utf8"),
      context,
    );
  return { containers, listeners, body, run };
}

describe("public form embeds", () => {
  it("refuses unexpected origins, protocols and paths", () => {
    const app = fixture([
      "https://attacker.example/f/test",
      "javascript:alert(1)",
      "https://forms.example/settings",
      "https://forms.example/f/test",
    ]);
    app.run();
    expect(app.containers.map((c) => c.children.length)).toEqual([0, 0, 0, 1]);
    const url = new URL(app.containers[3].children[0].src);
    expect(url.searchParams.get("utm_source")).toBe("newsletter");
    expect(url.searchParams.has("email")).toBe(false);
    expect(url.searchParams.get("embedOrigin")).toBe(
      "https://publisher.example",
    );
  });
  it("checks both source and origin and isolates multiple iframes", () => {
    const app = fixture([
      "https://forms.example/f/one",
      "https://forms.example/f/two",
    ]);
    app.run();
    const first = app.containers[0].children[0];
    const second = app.containers[1].children[0];
    app.listeners.forEach((fn) =>
      fn({
        origin: "https://attacker.example",
        source: first.contentWindow,
        data: { type: "xem:resize", height: 999 },
      }),
    );
    expect(first.style.height).toBeUndefined();
    app.listeners.forEach((fn) =>
      fn({
        origin: "https://forms.example",
        source: {},
        data: { type: "xem:resize", height: 999 },
      }),
    );
    expect(first.style.height).toBeUndefined();
    app.listeners.forEach((fn) =>
      fn({
        origin: "https://forms.example",
        source: first.contentWindow,
        data: { type: "xem:resize", height: 999 },
      }),
    );
    expect(first.style.height).toBe("999px");
    expect(second.style.height).toBeUndefined();
    app.listeners.forEach((fn) =>
      fn({
        origin: "https://forms.example",
        source: first.contentWindow,
        data: { type: "xem:resize", height: Infinity },
      }),
    );
    expect(first.style.height).toBe("999px");
  });
  it("does not duplicate mounts or listeners when included twice", () => {
    const app = fixture(["https://forms.example/f/one"]);
    app.run();
    app.run();
    expect(app.containers[0].children).toHaveLength(1);
    expect(app.listeners).toHaveLength(1);
  });
  it("renders legacy popup and sidebar placements as inset sheets and closes the previous sheet", () => {
    const app = fixture([
      "https://forms.example/f/one",
      "https://forms.example/f/two",
    ]);
    app.containers[0].setAttribute("data-xem-mode", "popup");
    app.containers[1].setAttribute("data-xem-mode", "sidebar");
    app.run();
    for (const sheet of app.body.children) {
      expect(sheet.style.cssText).toContain("inset:12px 12px 12px auto");
      expect(sheet.style.cssText).toContain("height:calc(100dvh - 24px)");
      expect(sheet.style.cssText).not.toContain("margin:auto");
    }
    app.containers[0].children[0].dispatchEvent({ type: "click" });
    expect(app.body.children[0].open).toBe(true);
    app.containers[1].children[0].dispatchEvent({ type: "click" });
    expect(app.body.children[0].open).toBe(false);
    expect(app.body.children[1].open).toBe(true);
    app.body.children[1].children[0].dispatchEvent({ type: "click" });
    expect(app.body.children[1].open).toBe(false);
  });
});


test("accepts the new sheet placement", () => {
  const app = fixture(["https://forms.example/f/one"]);
  app.containers[0].setAttribute("data-xem-mode", "sheet");
  app.run();
  expect(app.containers[0].children[0].tag).toBe("button");
  expect(app.body.children[0].tag).toBe("dialog");
});
