<script setup lang="ts">
import { MantineProvider, type MantineThemeOverride } from '@mantine-vue/core'

const theme: MantineThemeOverride = {
  fontSizes: {
    xxl: '1.5rem', // 24px (defaults: xs 12, sm 14, md 16, lg 18, xl 20)
  },
  lineHeights: {
    xxl: '1.7',
  },
}
</script>

<template>
  <MantineProvider default-color-scheme="auto" :theme="theme">
    <slot />
  </MantineProvider>
</template>
