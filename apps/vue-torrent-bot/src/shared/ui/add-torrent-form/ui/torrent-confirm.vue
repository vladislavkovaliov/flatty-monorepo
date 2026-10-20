<script setup lang="ts">
import { Alert, Button, Group, Stack, Text } from '@mantine-vue/core'

import {
  MEDIA_TYPE_LABELS,
  QUALITY_LABELS,
  savePath,
  type MediaType,
  type Quality,
} from '@/entities/task'

interface Props {
  name: string
  mediaType: MediaType
  quality: Quality
  testMode?: boolean
  submitting?: boolean
}

withDefaults(defineProps<Props>(), {
  testMode: true,
  submitting: false,
})

const emit = defineEmits<{ submit: []; back: []; cancel: [] }>()
</script>

<template>
  <Stack gap="xs">
    <Alert color="green" variant="light">Готово к добавлению</Alert>
    <Text size="sm">Название: {{ name }}</Text>
    <Text size="sm">Категория: {{ MEDIA_TYPE_LABELS[mediaType] }}</Text>
    <Text size="sm">Качество: {{ QUALITY_LABELS[quality] }}</Text>
    <Text size="sm">Путь: {{ savePath(mediaType) }}</Text>
    <Alert v-if="testMode" color="yellow" variant="light">
      TEST MODE — файл реально скачиваться не будет.
    </Alert>
    <Group>
      <Button :loading="submitting" @click="emit('submit')">Добавить в очередь</Button>
      <Button :disabled="submitting" variant="default" @click="emit('back')">Изменить</Button>
      <Button :disabled="submitting" color="red" variant="subtle" @click="emit('cancel')">
        Отмена
      </Button>
    </Group>
  </Stack>
</template>
