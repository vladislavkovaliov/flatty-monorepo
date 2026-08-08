import { execFile } from "node:child_process";
import path from "node:path";
import { promisify } from "node:util";

const execFileAsync = promisify(execFile);

export interface ParsedInvoice {
  vendor?: string;
  amount?: number;
  date?: string;
  description?: string;
  category?: string;
}

const INVOICE_PARSER_PATH = path.join(process.cwd(), "mcps", "invoice-parser");

export async function parseInvoiceFile(
  filePath: string,
): Promise<ParsedInvoice[]> {
  let stdout: string;
  let stderr: string;

  try {
    const result = await execFileAsync(
      INVOICE_PARSER_PATH,
      ["-file", filePath],
      { timeout: 30_000, maxBuffer: 10 * 1024 * 1024 },
    );
    stdout = result.stdout;
    stderr = result.stderr;
  } catch (error) {
    console.error("Invoice parser failed to run:", error);
    throw new Error("Invoice parser failed to run");
  }

  const trimmed = stdout.trim();

  if (trimmed === "" || trimmed === "null") {
    if (stderr.trim()) {
      console.error("Invoice parser stderr:", stderr);
    }
    throw new Error("Invoice parser returned no data");
  }

  try {
    const data: unknown = JSON.parse(trimmed);

    if (!Array.isArray(data)) {
      throw new Error("Invoice parser returned invalid data");
    }

    return data as ParsedInvoice[];
  } catch {
    throw new Error("Invoice parser returned invalid data");
  }
}
