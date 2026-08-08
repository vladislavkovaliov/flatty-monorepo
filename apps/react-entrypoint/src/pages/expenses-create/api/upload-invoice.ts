import type { ParsedInvoice } from "@/lib/parseInvoice";

export interface UploadInvoiceResult {
  parsed: ParsedInvoice[] | null;
  parseError: string | null;
}

export async function uploadInvoice(
  file: File,
  signal?: AbortSignal,
): Promise<UploadInvoiceResult> {
  const formData = new FormData();
  formData.append("file", file);

  const response = await fetch("/api/uploads", {
    method: "POST",
    body: formData,
    signal,
  });

  if (!response.ok) {
    const body = (await response.json().catch(() => null)) as {
      error?: string;
    } | null;
    throw new Error(body?.error ?? "Failed to upload the PDF");
  }

  return (await response.json()) as UploadInvoiceResult;
}
