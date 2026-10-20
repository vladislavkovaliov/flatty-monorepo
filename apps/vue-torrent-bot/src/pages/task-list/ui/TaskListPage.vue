<script setup lang="ts">
import { formatSize } from '@/entities/task'
import { useTaskList } from '../model/useTaskList'
import HeroCard from '@/shared/ui/hero-card/hero-card.vue'
import { Flex, Stack, Button, Text, Chip, Box } from '@mantine-vue/core'

defineEmits<{ open: [id: number] }>()

const { data, isPending, isError, error } = useTaskList()
</script>

<template>
  <section>
    <HeroCard>
      <template #header>
        <Text c="black">Active</Text>
      </template>
    </HeroCard>
  </section>
  <section>
    <h2>Мои загрузки</h2>

    <p v-if="isPending">Загрузка…</p>
    <p v-else-if="isError">Ошибка: {{ error?.message }}</p>
    <p v-else-if="data?.length === 0">Пока нет загрузок</p>

    <ul v-else>
      <li v-for="task in data" :key="task.id">
        <button type="button" @click="$emit('open', task.id)">
          {{ task.name }} — {{ task.status }} — {{ formatSize(task.size_bytes) }}
        </button>
      </li>
    </ul>
  </section>
</template>
