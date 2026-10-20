<script setup lang="ts">
import { formatSize, useTaskActions } from '@/entities/task'
import { useTask } from '../model/useTask'

const props = defineProps<{ taskId: number }>()

defineEmits<{ back: [] }>()

const { data, isPending, isError, error } = useTask(() => props.taskId)
const { pause, resume, cancel, remove } = useTaskActions()
</script>

<template>
  <section>
    <button type="button" @click="$emit('back')">Назад</button>
    <h2>Детали задачи</h2>

    <p v-if="isPending">Загрузка…</p>
    <p v-else-if="isError">Ошибка: {{ error?.message }}</p>

    <div v-else-if="data">
      <p>Название: {{ data.name }}</p>
      <p>Статус: {{ data.status }}</p>
      <p>Прогресс: {{ data.progress }}</p>
      <p>Размер: {{ formatSize(data.size_bytes) }}</p>
      <p>Magnet: {{ data.magnet }}</p>

      <button type="button" @click="pause.mutate(data.id)">Пауза</button>
      <button type="button" @click="resume.mutate(data.id)">Продолжить</button>
      <button type="button" @click="cancel.mutate(data.id)">Отменить</button>
      <button type="button" @click="remove.mutate(data.id)">Удалить</button>
    </div>
  </section>
</template>
