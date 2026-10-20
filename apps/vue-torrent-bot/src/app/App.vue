<script setup lang="ts">
import { ref } from 'vue'
import type { IAppConfig } from '@flatty-budget/mfe-host'
import { Box, Flex, Paper } from '@mantine-vue/core'
import HelpPage from '@/pages/help/ui/HelpPage.vue'
import SettingsPage from '@/pages/settings/ui/SettingsPage.vue'
import TaskAddPage from '@/pages/task-add/ui/TaskAddPage.vue'
import TaskDetailPage from '@/pages/task-detail/ui/TaskDetailPage.vue'
import TaskListPage from '@/pages/task-list/ui/TaskListPage.vue'
import AppMenu from '@/shared/ui/app-menu/app-menu.vue'

defineProps<{ config: IAppConfig }>()

const NAV_ITEMS = [
  { key: 'tasks', label: 'Мои загрузки' },
  { key: 'add', label: 'Добавить' },
  { key: 'settings', label: 'Настройки' },
  { key: 'help', label: 'Помощь' },
] as const

const PAGE_NAMES = ['tasks', 'add', 'detail', 'settings', 'help'] as const

type PageName = (typeof PAGE_NAMES)[number]

const page = ref<PageName>('tasks')
const selectedTaskId = ref<number | null>(null)

function isPageName(value: string): value is PageName {
  return (PAGE_NAMES as readonly string[]).includes(value)
}

function select(key: string) {
  if (isPageName(key)) {
    page.value = key
  }
}

function openTask(id: number) {
  selectedTaskId.value = id
  page.value = 'detail'
}
</script>

<template>
  <Flex align="flex-start" gap="md">
    <Paper with-border p="xs" :w="240">
      <AppMenu :items="NAV_ITEMS" :model-value="page" @update:model-value="select" />
    </Paper>

    <Box flex="1">
      <TaskListPage v-if="page === 'tasks'" @open="openTask" />
      <TaskAddPage v-else-if="page === 'add'" @done="page = 'tasks'" />
      <TaskDetailPage
        v-else-if="page === 'detail' && selectedTaskId !== null"
        :task-id="selectedTaskId"
        @back="page = 'tasks'"
      />
      <SettingsPage v-else-if="page === 'settings'" />
      <HelpPage v-else-if="page === 'help'" />
    </Box>
  </Flex>
</template>
