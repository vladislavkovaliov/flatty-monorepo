<script setup lang="ts">
import { Flex, Stack, Text, Badge } from '@mantine-vue/core'
import type { MediaType, Quality, TorrentStatus } from '@/entities/task/index.ts'

interface Props {
  name: string
  media_type: MediaType
  quality: Quality
  status: TorrentStatus
}

defineProps<Props>()
</script>

<template>
  <Flex>
    <Stack gap="xs">
      <Text>{{ name }}</Text>
      <Flex gap="xs">
        <Badge>{{ media_type }}</Badge>
        <Badge>{{ quality }}</Badge>
      </Flex>
    </Stack>
  </Flex>
</template>
