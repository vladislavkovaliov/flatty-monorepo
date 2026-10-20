<script setup lang="ts">
import { Box, Chip, Flex, Stack, Text, Badge } from '@mantine-vue/core'

interface Props {
  title?: string
  mediaType?: string
  quality?: string
  status?: string
  progress?: number
}

withDefaults(defineProps<Props>(), {
  title: 'Matrix.1999.1080p.Bluray',
  mediaType: 'Movie',
  quality: '1080p',
  status: 'in progress',
})
</script>

<template>
  <Flex>
    <Stack gap="xs">
      <Text>{{ title }}</Text>
      <Flex gap="xs">
        <Badge>{{ mediaType }}</Badge>
        <Badge>{{ quality }}</Badge>
      </Flex>
    </Stack>
  </Flex>
</template>
