<script setup lang="ts">
import type { AddTaskInput } from '@/entities/task'
import AddTorrentForm from '@/shared/ui/add-torrent-form/add-torrent-form.vue'

import { useAddTask } from '../model/useAddTask'

const emit = defineEmits<{ done: [] }>()

const { mutate, isPending, error } = useAddTask()

function submit(payload: AddTaskInput) {
  mutate(payload, { onSuccess: () => emit('done') })
}
</script>

<template>
  <AddTorrentForm
    :error="error ? error.message : null"
    :submitting="isPending"
    @cancel="emit('done')"
    @submit="submit"
  />
</template>
