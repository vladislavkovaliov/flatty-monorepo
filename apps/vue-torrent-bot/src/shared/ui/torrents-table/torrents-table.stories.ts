import type { Meta, StoryObj } from '@storybook/vue3-vite'

import TorrentsTable from './torrents-table.vue'

const meta = {
  title: 'shared/ui/TorrentsTable',
  component: TorrentsTable,
  args: {
    items: [
      {
        id: 1,
        name: 'Matrix.1999.1080p.Bluray',
        media_type: 'movie',
        quality: '1080p',
        status: 'downloading',
        progress: 30,
      },
      {
        id: 2,
        name: 'Dune.2021.2160p.WEB-DL',
        media_type: 'movie',
        quality: '2160p',
        status: 'completed',
        progress: 100,
      },
      {
        id: 3,
        name: 'Severance.S02.1080p.WEBRip',
        media_type: 'series',
        quality: '1080p',
        status: 'paused',
        progress: 90,
      },
    ],
  },
} satisfies Meta<typeof TorrentsTable>

export default meta
type Story = StoryObj<typeof meta>

export const Default: Story = {}
export const Empty: Story = { args: { items: [] } }
