<script setup lang="ts">
import { Button, Code, CopyButton, Group, Stack, Text, TextInput } from '@mantine-vue/core'

import { formatSize, type MagnetInfo } from '@/entities/task'

interface Props {
  label?: string
  error?: string | null
  parsed?: MagnetInfo | null
  disabled?: boolean
}

withDefaults(defineProps<Props>(), {
  label: 'Magnet-ссылка',
  error: null,
  parsed: null,
  disabled: false,
})

const model = defineModel<string>({ required: true })

const emit = defineEmits<{ check: [] }>()
</script>

<template>
  <Stack gap="xs">
    <TextInput
      v-model="model"
      :disabled="disabled"
      :error="error"
      :label="label"
      placeholder="magnet:?xt=urn:btih:..."
    />
    <Group>
      <Button :disabled="disabled || !model" @click="emit('check')">Проверить</Button>
    </Group>

    <Stack v-if="parsed" gap="xs">
      <Text fw="{600}">Torrent найден</Text>
      <Text size="sm">Название: {{ parsed.displayName ?? 'не определён' }}</Text>
      <Group gap="xs">
        <Text size="sm">
          Hash: <Code>{{ parsed.infoHash }}</Code>
        </Text>
        <CopyButton :value="parsed.infoHash">
          <template #default="{ copied, copy }">
            <Button size="xs" variant="subtle" @click="copy()">
              {{ copied ? 'Скопировано' : 'Копировать' }}
            </Button>
          </template>
        </CopyButton>
      </Group>
      <Text size="sm">Trackers: {{ parsed.trackers.length }}</Text>
      <Text size="sm">Размер: {{ formatSize(parsed.sizeBytes) }}</Text>
    </Stack>
  </Stack>
</template>
