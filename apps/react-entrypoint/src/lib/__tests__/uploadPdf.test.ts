import { mkdtemp, readdir, rm } from "node:fs/promises";
import os from "node:os";
import path from "node:path";

import { describe, expect, it } from "vitest";

import {
  buildPdfName,
  MAX_PDF_SIZE_BYTES,
  PdfUploadError,
  savePdf,
} from "../uploadPdf";

function makeFile(name: string, type: string, size = 3): File {
  return new File([new Uint8Array(size)], name, { type });
}

describe("buildPdfName", () => {
  it("returns a name ending with .pdf", () => {
    expect(buildPdfName()).toMatch(/\.pdf$/);
  });

  it("returns unique names", () => {
    expect(buildPdfName()).not.toBe(buildPdfName());
  });
});

describe("savePdf", () => {
  it("rejects an empty file", async () => {
    await expect(
      savePdf(makeFile("a.pdf", "application/pdf", 0), "/tmp"),
    ).rejects.toBeInstanceOf(PdfUploadError);
  });

  it("rejects a non-PDF file", async () => {
    await expect(
      savePdf(makeFile("a.txt", "text/plain"), "/tmp"),
    ).rejects.toThrow("Only PDF files are allowed");
  });

  it("rejects an oversized file", async () => {
    const oversized = new Uint8Array(MAX_PDF_SIZE_BYTES + 1);
    await expect(
      savePdf(
        new File([oversized], "big.pdf", { type: "application/pdf" }),
        "/tmp",
      ),
    ).rejects.toThrow("10 MB limit");
  });

  it("writes a valid PDF to the target directory and returns its filename", async () => {
    const dir = await mkdtemp(path.join(os.tmpdir(), "pdf-test-"));
    try {
      const filename = await savePdf(
        makeFile("receipt.pdf", "application/pdf"),
        dir,
      );

      expect(filename).toMatch(/\.pdf$/);

      const files = await readdir(dir);
      expect(files).toContain(filename);
    } finally {
      await rm(dir, { recursive: true, force: true });
    }
  });
});
