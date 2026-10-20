<script setup lang="ts">
import { SegmentedControl, Stack, Text } from '@mantine-vue/core'

import { QUALITY_OPTIONS, type Quality } from '@/entities/task'

const model = defineModel<Quality>({ required: true })
</script>

<template>
  <Stack gap="xs">
    <Text size="sm" fw="{500}">Выберите качество:</Text>
    <SegmentedControl v-model="model" :data="QUALITY_OPTIONS" />
  </Stack>
</template>
