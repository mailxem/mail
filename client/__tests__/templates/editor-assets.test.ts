import { prepareEditorAssets, restoreEditorAssets } from "@/lib/template-starters/editor-assets";

describe("local editor asset adaptation", () => {
  const signal = new AbortController().signal;
  it("uses same-origin raster bytes in the local iframe and restores export URLs", async () => {
    const url = "http://localhost:3000/assets/template-starters/media/logo.png";
    global.fetch = jest.fn().mockResolvedValue({ ok: true, headers: new Headers({ "content-type": "image/png" }), arrayBuffer: async () => new Uint8Array([1, 2, 3]).buffer });
    const original = { image: url, userImage: "https://example.com/photo.png" };
    const prepared = await prepareEditorAssets(original, signal, "http://localhost:3000");
    expect(prepared.design.image).toBe("data:image/png;base64,AQID");
    expect(original.image).toBe(url);
    expect(global.fetch).toHaveBeenCalledWith(url, { signal });
    const restored = restoreEditorAssets({ design: prepared.design, html: `<img src="${prepared.design.image}">` }, prepared.sources);
    expect(restored.design).toEqual(original);
    expect(restored.html).toBe(`<img src="${url}">`);
  });
  it("does not fetch or rewrite production or customer assets", async () => {
    global.fetch = jest.fn();
    const design = { image: "https://xem.email/assets/template-starters/media/logo.png" };
    expect((await prepareEditorAssets(design, signal, "https://xem.email")).design).toEqual(design);
    expect(global.fetch).not.toHaveBeenCalled();
  });
  it("rejects active image formats", async () => {
    global.fetch = jest.fn().mockResolvedValue({ ok: true, headers: new Headers({ "content-type": "image/svg+xml" }) });
    await expect(prepareEditorAssets({ image: "http://localhost:3000/assets/template-starters/x.svg" }, signal, "http://localhost:3000")).rejects.toThrow("Unsupported");
  });
});
