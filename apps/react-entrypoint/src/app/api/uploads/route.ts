import path from "node:path";

import { NextResponse } from "next/server";

import { type ParsedInvoice, parseInvoiceFile } from "@/lib/parseInvoice";
import { MAX_PDF_SIZE_BYTES, PdfUploadError, savePdf } from "@/lib/uploadPdf";

const AUTH_API_URL = process.env.AUTH_API_URL ?? "http://localhost:3000";

async function isAuthenticated(request: Request): Promise<boolean> {
  const cookie = request.headers.get("cookie");

  if (!cookie) {
    return false;
  }

  try {
    const response = await fetch(`${AUTH_API_URL}/api/auth/get-session`, {
      headers: { cookie },
      cache: "no-store",
    });

    if (!response.ok) {
      return false;
    }

    const data = (await response.json()) as { session?: unknown };

    return data.session != null;
  } catch {
    return false;
  }
}

export async function POST(request: Request) {
  if (!(await isAuthenticated(request))) {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
  }

  // Reject oversized bodies before buffering them into memory. Multipart
  // overhead adds a small margin on top of the raw file size.
  const contentLength = Number(request.headers.get("content-length") ?? 0);
  if (contentLength > MAX_PDF_SIZE_BYTES + 1024) {
    return NextResponse.json(
      { error: "File exceeds the 10 MB limit" },
      { status: 413 },
    );
  }

  try {
    let formData: FormData;

    try {
      formData = await request.formData();
    } catch {
      return NextResponse.json(
        { error: "Expected multipart form data" },
        { status: 400 },
      );
    }

    const file = formData.get("file");

    if (!(file instanceof File)) {
      return NextResponse.json({ error: "No file uploaded" }, { status: 400 });
    }

    const filename = await savePdf(file);

    const filePath = path.join(process.cwd(), "pdfs", filename);
    let parsed: ParsedInvoice[] | null = null;
    let parseError: string | null = null;

    try {
      parsed = await parseInvoiceFile(filePath);
    } catch (error) {
      parseError =
        error instanceof Error ? error.message : "Failed to parse the invoice";
    }

    return NextResponse.json({ filename, parsed, parseError });
  } catch (error) {
    if (error instanceof PdfUploadError) {
      return NextResponse.json(
        { error: error.message },
        { status: error.status },
      );
    }

    console.error("PDF upload failed:", error);
    return NextResponse.json(
      { error: "Failed to save the file" },
      { status: 500 },
    );
  }
}
