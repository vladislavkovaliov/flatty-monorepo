import React, { Suspense } from 'react';
import { Center, Loader } from '@mantine/core';

export function lazyLoad<T extends React.ComponentType<any>>(
  importFn: () => Promise<{ default: T }>,
) {
  const LazyComponent = React.lazy(importFn);
  
  return (props: React.ComponentProps<T>) => (
    <Suspense fallback={<Center h="50vh"><Loader /></Center>}>
      <LazyComponent {...props} />
    </Suspense>
  );
}