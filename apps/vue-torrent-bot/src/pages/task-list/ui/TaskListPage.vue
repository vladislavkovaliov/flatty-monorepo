<script setup lang="ts">
import { useTaskList } from '../model/useTaskList'
import { useDownloadingTasks } from '../model/useDownloadingTasks'
import HeroCard from '@/shared/ui/hero-card/hero-card.vue'
import TorrentsTable from '@/shared/ui/torrents-table/torrents-table.vue'
import { Text } from '@mantine-vue/core'

defineEmits<{ open: [id: number] }>()

const { data, isPending, isError, error } = useTaskList()
const { data: downloadingCount, isPending: isDownloadingCountPending } = useDownloadingTasks()

</script>

<template>
  <section v-show="!isDownloadingCountPending">
    <HeroCard type="download" :text="downloadingCount">
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

    <TorrentsTable v-else :items="data" />
  </section>
</template>
