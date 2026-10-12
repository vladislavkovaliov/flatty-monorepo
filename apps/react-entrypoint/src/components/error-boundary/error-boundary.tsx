"use client";

import { Box, Button, Text, Title } from "@mantine/core";
import { Component, type ErrorInfo, type ReactNode } from "react";

interface Props {
  children: ReactNode;
  /** Optional fallback label shown on the retry button. */
  fallbackLabel?: string;
}

interface State {
  hasError: boolean;
  error: Error | null;
}

/**
 * Minimal class-based error boundary.
 * Catches render errors from child components (e.g. remote micro-frontends)
 * and shows a fallback UI with a retry button.
 */
export class ErrorBoundary extends Component<Props, State> {
  state: State = { hasError: false, error: null };

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, info: ErrorInfo) {
    console.error("[ErrorBoundary]", error, info.componentStack);
  }

  handleRetry = () => {
    this.setState({ hasError: false, error: null });
  };

  render() {
    if (this.state.hasError) {
      return (
        <Box p="xl" ta="center">
          <Title order={4}>Something went wrong</Title>
          <Text c="dimmed" mt="xs" size="sm">
            {this.state.error?.message ?? "An unexpected error occurred."}
          </Text>
          <Button mt="md" variant="light" onClick={this.handleRetry}>
            {this.props.fallbackLabel ?? "Try again"}
          </Button>
        </Box>
      );
    }

    return this.props.children;
  }
}
