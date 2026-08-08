import { existsSync } from "node:fs";
import { mkdtemp, rm, writeFile } from "node:fs/promises";
import os from "node:os";
import path from "node:path";

import { describe, expect, it } from "vitest";

import { parseInvoiceFile } from "../parseInvoice";

function buildValidPdf(): string {
  const objects = [
    "<< /Type /Catalog /Pages 2 0 R >>",
    "<< /Type /Pages /Kids [3 0 R] /Count 1 >>",
    "<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] >>",
  ];

  let content = "%PDF-1.4\n";
  const offsets = [0];

  for (const [index, object] of objects.entries()) {
    offsets.push(Buffer.byteLength(content, "latin1"));
    content += `${index + 1} 0 obj\n${object}\nendobj\n`;
  }

  const xrefStart = Buffer.byteLength(content, "latin1");
  content += `xref\n0 ${objects.length + 1}\n0000000000 65535 f \n`;

  for (let i = 1; i <= objects.length; i += 1) {
    content += `${String(offsets[i]).padStart(10, "0")} 00000 n \n`;
  }

  content += `trailer\n<< /Size ${objects.length + 1} /Root 1 0 R >>\nstartxref\n${xrefStart}\n%%EOF\n`;
  return content;
}

describe("parseInvoiceFile", () => {
  it.skipIf(!existsSync(path.join(process.cwd(), "mcps", "invoice-parser")))(
    "parses a valid PDF and returns invoice data",
    async () => {
      const dir = await mkdtemp(path.join(os.tmpdir(), "invoice-test-"));
      try {
        const filePath = path.join(dir, "invoice.pdf");
        await writeFile(filePath, buildValidPdf());

        const result = await parseInvoiceFile(filePath);

        expect(Array.isArray(result)).toBe(true);
        expect(result.length).toBeGreaterThan(0);
        expect(typeof result[0].vendor).toBe("string");
      } finally {
        await rm(dir, { recursive: true, force: true });
      }
    },
    30_000,
  );

  it("throws when the file does not exist", async () => {
    await expect(
      parseInvoiceFile("/nonexistent/invoice.pdf"),
    ).rejects.toThrow();
  });
});
