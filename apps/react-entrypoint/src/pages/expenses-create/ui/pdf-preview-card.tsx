import { Card, Text } from "@mantine/core";

interface PdfPreviewCardProps {
  url: string | null;
}

export function PdfPreviewCard({ url }: PdfPreviewCardProps) {
  return (
    <Card withBorder padding="md">
      <Text fw={600} mb="sm">
        Receipt Preview
      </Text>
      {url ? (
        <iframe
          src={url}
          title="Receipt PDF"
          style={{ width: "100%", height: 600, border: "none" }}
        />
      ) : (
        <Text size="sm" c="dimmed">
          Select a PDF to preview it here
        </Text>
      )}
    </Card>
  );
}
