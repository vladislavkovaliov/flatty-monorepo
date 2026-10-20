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

withDefaults(
  defineProps<{
    items?: {
      title?: string
      mediaType?: string
      quality?: string
      status?: string
      progress?: number
    }[]
  }>(),
  { items: () => [] },
)
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
      <TableTr v-for="(item, index) in items" :key="index">
        <TableTd>
          <TorrentItem v-bind="item" />
        </TableTd>
        <TableTd>{{ item.status }}</TableTd>
        <TableTd>
          <Progress :value="item.progress" />
        </TableTd>
      </TableTr>
    </TableTbody>
  </Table>
</template>
