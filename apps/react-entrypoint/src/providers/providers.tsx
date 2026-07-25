"use client";

import { SingleTabManagerWrapper } from "@/components/single-tab-manager";
import { QueryProvider } from "./query-provider";
import { ThemeProvider } from "./theme-provider";

export function Providers({ children }: { children: React.ReactNode }) {
  return (
    <SingleTabManagerWrapper>
      <QueryProvider>
        <ThemeProvider>{children}</ThemeProvider>
      </QueryProvider>
    </SingleTabManagerWrapper>
  );
}
