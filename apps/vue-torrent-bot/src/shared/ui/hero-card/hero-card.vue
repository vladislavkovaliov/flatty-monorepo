<script setup lang="ts">
  import { Flex, Stack, Text, Box } from '@mantine-vue/core'
  import { PhCheckCircle, PhDownload, PhGauge } from '@phosphor-icons/vue'

  import DownloadTextSpeed from './ui/download-text-speed.vue'

  defineProps<{
    type?: 'download' | 'done' | 'speed'
    text?: string
  }>()
</script>

<style scoped>
  .container {
    max-width: 243px;
    padding: 16px;
    width: 100%;
    border: 1px solid #dee2e6;
    border-radius: 8px;
    background: #fff;
  }

  .icon {
    width: 28px;
    height: 28px;
    background: #e7f5ff;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 4px;
  }
</style>

<template>

  <Stack class="container" gap="xs">
    <Flex justify="space-between">
      <slot name="header" />

      <Box v-if="type" class="icon">
        <PhDownload v-if="type === 'download'" :size="16" color="#1971c2" />
        <PhCheckCircle v-if="type === 'done'" :size="16" />
        <PhGauge v-if="type === 'speed'" :size="16" />
      </Box>

    </Flex>

    <Flex>
      <Text v-if="type === 'done' || type === 'download'" size="xxl">{{ text }}</Text>
      <DownloadTextSpeed v-if="type === 'speed'" :text="text" speed="MB/s" :showArrow="true" />
    </Flex>

    <Flex>
      <Text c="dimmed">Active now</Text>
    </Flex>
  </Stack>

</template>
