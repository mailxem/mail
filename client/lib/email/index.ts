export const createTempFile = async (
  content: string,
  filename: string
): Promise<File> => {
  if (typeof window !== "undefined") {
    throw new Error("createTempFile is not supported in the browser");
  }

  const fs = require("node:fs");
  const path = require("node:path");
  const os = require("node:os");

  // Create temp file path
  const tempPath = path.join(os.tmpdir(), filename);

  // Write content to temp file
  await fs.promises.writeFile(tempPath, content, "utf8");

  // Create File object from temp file
  const tempFile = new File([await fs.promises.readFile(tempPath)], filename, {
    type: "text/html",
  });

  // Clean up temp file
  await fs.promises.unlink(tempPath);

  return tempFile;
};
