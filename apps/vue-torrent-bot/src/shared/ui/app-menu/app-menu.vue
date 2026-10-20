<script setup lang="ts">
import { NavLink, ScrollArea, Stack } from '@mantine-vue/core'

const props = defineProps<{ items: readonly { key: string; label: string }[] }>()

const model = defineModel<string>({ required: true })
</script>

<template>
  <ScrollArea>
    <Stack gap="xs">
      <NavLink
        v-for="item in props.items"
        :key="item.key"
        :active="item.key === model"
        :label="item.label"
        component="button"
        variant="light"
        @click="model = item.key"
      />
    </Stack>
  </ScrollArea>
</template>
