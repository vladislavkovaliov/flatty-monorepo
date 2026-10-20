<script setup lang="ts">
import { ref } from 'vue'
import { Alert, Button, Card, Divider, Group, Stack } from '@mantine-vue/core'

import {
  parseMagnet,
  type AddTaskInput,
  type MagnetInfo,
  type MediaType,
  type Quality,
} from '@/entities/task'

import CategoryPicker from './ui/category-picker.vue'
import MagnetField from './ui/magnet-field.vue'
import QualityPicker from './ui/quality-picker.vue'
import TorrentConfirm from './ui/torrent-confirm.vue'

type Step = 'input' | 'details' | 'confirm'

interface Props {
  testMode?: boolean
  submitting?: boolean
  error?: string | null
  initialCategory?: MediaType
  initialQuality?: Quality
  initialMagnet?: string
  /** Storybook/testing helper: open the form on a specific step. */
  initialStep?: Step
}

const props = withDefaults(defineProps<Props>(), {
  testMode: true,
  submitting: false,
  error: null,
  initialCategory: 'movie',
  initialQuality: '1080p',
  initialMagnet: '',
})

const emit = defineEmits<{ submit: [payload: AddTaskInput]; cancel: [] }>()

const magnet = ref(props.initialMagnet)
const parsed = ref<MagnetInfo | null>(parseMagnet(props.initialMagnet))
const magnetError = ref<string | null>(null)
const category = ref<MediaType>(props.initialCategory)
const quality = ref<Quality>(props.initialQuality)
const step = ref<Step>(parsed.value ? (props.initialStep ?? 'details') : 'input')

function check() {
  parsed.value = parseMagnet(magnet.value)
  magnetError.value = parsed.value ? null : 'Ссылка не распознана'
  if (parsed.value) step.value = 'details'
}

function submit() {
  emit('submit', {
    magnet: magnet.value,
    media_type: category.value,
    quality: quality.value,
  })
}
</script>

<template>
  <Card padding="md" with-border>
    <Stack gap="md">
      <MagnetField
        v-model="magnet"
        :disabled="props.submitting"
        :error="magnetError"
        :parsed="parsed"
        @check="check"
      />

      <template v-if="step === 'details' && parsed">
        <Divider />
        <CategoryPicker v-model="category" />
        <QualityPicker v-model="quality" />
        <Group>
          <Button @click="step = 'confirm'">Далее</Button>
          <Button variant="subtle" @click="emit('cancel')">Отмена</Button>
        </Group>
      </template>

      <template v-else-if="step === 'confirm' && parsed">
        <Divider />
        <TorrentConfirm
          :media-type="category"
          :name="parsed.displayName ?? parsed.infoHash"
          :quality="quality"
          :submitting="props.submitting"
          :test-mode="props.testMode"
          @back="step = 'details'"
          @cancel="emit('cancel')"
          @submit="submit"
        />
      </template>

      <Alert v-if="props.error" color="red" variant="light">{{ props.error }}</Alert>
    </Stack>
  </Card>
</template>
