"use client";

import { AppShell } from "@mantine/core";
import { useDisclosure } from "@mantine/hooks";
import { Spotlight } from "@mantine/spotlight";
import { useRouter } from "next/navigation";
import { AppShellHeader } from "@/components/app-shell-header";
import { AuthGuard } from "@/features/auth/guard/auth-guard";
import { buildSpotlightActions } from "@/features/spotlight";
import { APPS, type AppDefinition } from "@/shared/config/apps";

export default function MainLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  const [opened, { toggle }] = useDisclosure();
  const router = useRouter();

  const handleNavigate = (app: AppDefinition) => router.push(app.path);
  const spotlightActions = buildSpotlightActions(APPS, handleNavigate);

  return (
    <AuthGuard>
      <AppShell header={{ height: 60 }}>
        <AppShellHeader opened={opened} toggle={toggle} />
        <AppShell.Main>{children}</AppShell.Main>
      </AppShell>

      <Spotlight
        shortcut="mod + K"
        actions={spotlightActions}
        nothingFound="Nothing found..."
        highlightQuery
        searchProps={{
          placeholder: "Search applications...",
        }}
      />
    </AuthGuard>
  );
}
