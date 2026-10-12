import type { Metadata } from "next";
import { Providers } from "@/providers/providers";

import "@mantine/core/styles.css";
import "@mantine/dates/styles.css";
import "@mantine/charts/styles.css";
import "@mantine/spotlight/styles.css";

export const metadata: Metadata = {
  title: "Flatty Budget",
  description: "Personal budget management application",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body>
        <Providers>{children}</Providers>
      </body>
    </html>
  );
}
