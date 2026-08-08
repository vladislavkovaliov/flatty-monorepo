import { randomUUID } from "node:crypto";
import { mkdir, writeFile } from "node:fs/promises";
import path from "node:path";

export const MAX_PDF_SIZE_BYTES = 10 * 1024 * 1024; // 10 MB

export class PdfUploadError extends Error {
  readonly status: number;

  constructor(message: string, status = 400) {
    super(message);
    this.name = "PdfUploadError";
    this.status = status;
  }
}

function isPdf(file: File): boolean {
  return (
    file.type === "application/pdf" || file.name.toLowerCase().endsWith(".pdf")
  );
}

export function buildPdfName(): string {
  return `${Date.now()}-${randomUUID()}.pdf`;
}

export async function savePdf(
  file: File,
  dir = path.join(process.cwd(), "pdfs"),
): Promise<string> {
  if (!file || file.size === 0) {
    throw new PdfUploadError("No file provided");
  }

  if (!isPdf(file)) {
    throw new PdfUploadError("Only PDF files are allowed");
  }

  if (file.size > MAX_PDF_SIZE_BYTES) {
    throw new PdfUploadError("File exceeds the 10 MB limit", 413);
  }

  const filename = buildPdfName();
  const buffer = Buffer.from(await file.arrayBuffer());

  await mkdir(dir, { recursive: true });
  await writeFile(path.join(dir, filename), buffer);

  return filename;
}
