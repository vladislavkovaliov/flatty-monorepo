"use client";

import { useEffect, useRef, useState } from "react";

import type { ParsedInvoice } from "@/lib/parseInvoice";

import { uploadInvoice } from "../api/upload-invoice";

interface UseInvoiceImportOptions {
  onParsed: (parsed: ParsedInvoice[]) => boolean;
  onPdfUrlChange: (url: string | null) => void;
}

export function useInvoiceImport({
  onParsed,
  onPdfUrlChange,
}: UseInvoiceImportOptions) {
  const [uploading, setUploading] = useState(false);
  const [parsedApplied, setParsedApplied] = useState(false);
  const [uploadError, setUploadError] = useState<string | null>(null);
  const [pdfUrl, setPdfUrl] = useState<string | null>(null);
  const pdfUrlRef = useRef<string | null>(null);
  const uploadSeqRef = useRef(0);
  const uploadAbortRef = useRef<AbortController | null>(null);

  useEffect(() => {
    return () => {
      if (pdfUrlRef.current) {
        URL.revokeObjectURL(pdfUrlRef.current);
      }
    };
  }, []);

  const setPdfPreview = (url: string | null) => {
    if (pdfUrlRef.current) {
      URL.revokeObjectURL(pdfUrlRef.current);
    }
    pdfUrlRef.current = url;
    setPdfUrl(url);
    onPdfUrlChange(url);
  };

  const handleFileChange = async (file: File | null) => {
    uploadAbortRef.current?.abort();
    setUploadError(null);
    setParsedApplied(false);

    if (!file) {
      setPdfPreview(null);
      return;
    }

    const seq = ++uploadSeqRef.current;
    const controller = new AbortController();
    uploadAbortRef.current = controller;

    setPdfPreview(URL.createObjectURL(file));
    setUploading(true);
    try {
      const body = await uploadInvoice(file, controller.signal);

      if (seq !== uploadSeqRef.current) return;

      if (body.parseError) {
        setUploadError(body.parseError);
      } else if (body.parsed && body.parsed.length > 0) {
        if (onParsed(body.parsed)) {
          setParsedApplied(true);
        }
      } else {
        setUploadError("No invoice data could be extracted from the PDF");
      }
    } catch (error) {
      if (seq === uploadSeqRef.current) {
        setUploadError(
          error instanceof Error ? error.message : "Failed to upload the PDF",
        );
      }
    } finally {
      if (seq === uploadSeqRef.current) {
        setUploading(false);
      }
    }
  };

  return { uploading, uploadError, parsedApplied, pdfUrl, handleFileChange };
}
