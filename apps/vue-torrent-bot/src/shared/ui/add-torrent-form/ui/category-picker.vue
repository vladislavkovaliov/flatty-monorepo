<script setup lang="ts">
import { SegmentedControl, Stack, Text } from '@mantine-vue/core'

import { MEDIA_TYPE_OPTIONS, type MediaType } from '@/entities/task'

const model = defineModel<MediaType>({ required: true })
</script>

<template>
  <Stack gap="xs">
    <Text size="sm" fw="{500}">Куда сохранить?</Text>
    <SegmentedControl v-model="model" :data="MEDIA_TYPE_OPTIONS" />
  </Stack>
</template>
