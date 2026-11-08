<script setup lang="ts">
import {
  Table,
  TableTbody,
  TableTd,
  TableTh,
  TableThead,
  TableTr,
  Progress,
} from '@mantine-vue/core'

import TorrentItem from './ui/torrent-item.vue'
import type { TorrentTask } from '@/entities/task/index.ts'

type TorrentRow = Pick<
  TorrentTask,
  'id' | 'name' | 'status' | 'progress' | 'media_type' | 'quality'
>

withDefaults(defineProps<{ items?: TorrentRow[] }>(), {
  items: () => [],
})
</script>

<template>
  <Table>
    <TableThead>
      <TableTr>
        <TableTh>Task</TableTh>
        <TableTh>Status</TableTh>
        <TableTh>Progress</TableTh>
      </TableTr>
    </TableThead>
    <TableTbody>
      <TableTr v-for="item in items" :key="item.id">
        <TableTd>
          <TorrentItem
            :name="item.name"
            :media_type="item.media_type"
            :quality="item.quality"
            :status="item.status"
          />
        </TableTd>
        <TableTd>{{ item.status }}</TableTd>
        <TableTd>
          <Progress :value="item.progress" />
        </TableTd>
      </TableTr>
    </TableTbody>
  </Table>
</template>
