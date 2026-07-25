"use client";

import {
  Alert,
  Button,
  Center,
  Paper,
  PasswordInput,
  Stack,
  Text,
  Title,
} from "@mantine/core";
import { useRouter } from "next/navigation";
import { useState } from "react";

export default function SetPasswordPage() {
  const router = useRouter();
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [isPending, setIsPending] = useState(false);

  const handleSubmit = async () => {
    setError(null);

    if (!newPassword) {
      setError("Password is required");
      return;
    }

    if (newPassword.length < 8) {
      setError("Password must be at least 8 characters");
      return;
    }

    if (newPassword !== confirmPassword) {
      setError("Passwords do not match");
      return;
    }

    setIsPending(true);
    try {
      const response = await fetch("/api/auth/set-password", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ newPassword }),
        credentials: "include",
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.error ?? "Failed to set password");
      }

      router.push("/");
    } catch (err) {
      setError(
        err instanceof Error
          ? err.message
          : "Something went wrong. Please try again.",
      );
    } finally {
      setIsPending(false);
    }
  };

  const handleSkip = () => {
    localStorage.setItem("skip-set-password", "true");
    router.push("/");
  };

  return (
    <Center h="100vh">
      <Paper withBorder shadow="md" p="xl" radius="md" miw={400}>
        <Stack gap="md">
          <Title order={2}>Set Your Password</Title>
          <Text c="dimmed">
            You&apos;re logged in via magic link. Set a password so you can sign
            in faster next time.
          </Text>

          {error && (
            <Alert color="red" variant="light" title="Error">
              {error}
            </Alert>
          )}

          <PasswordInput
            label="New Password"
            placeholder="Enter your password"
            value={newPassword}
            onChange={(e) => setNewPassword(e.currentTarget.value)}
            disabled={isPending}
            required
          />

          <PasswordInput
            label="Confirm Password"
            placeholder="Confirm your password"
            value={confirmPassword}
            onChange={(e) => setConfirmPassword(e.currentTarget.value)}
            disabled={isPending}
            required
          />

          <Button onClick={handleSubmit} loading={isPending} fullWidth>
            Set Password
          </Button>

          <Button
            variant="transparent"
            onClick={handleSkip}
            ta="center"
            c="dimmed"
            size="sm"
          >
            Skip for now
          </Button>
        </Stack>
      </Paper>
    </Center>
  );
}
