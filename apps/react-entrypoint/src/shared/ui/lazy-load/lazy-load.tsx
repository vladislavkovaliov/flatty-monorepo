import { Center, Loader } from "@mantine/core";
import React, { Suspense } from "react";

// biome-ignore lint/suspicious/noExplicitAny: generic lazy wrapper needs to accept any component
export function lazyLoad<T extends React.ComponentType<any>>(
  importFn: () => Promise<{ default: T }>,
) {
  const LazyComponent = React.lazy(importFn);

  return (props: React.ComponentProps<T>) => (
    <Suspense
      fallback={
        <Center h="50vh">
          <Loader />
        </Center>
      }
    >
      <LazyComponent {...props} />
    </Suspense>
  );
}
