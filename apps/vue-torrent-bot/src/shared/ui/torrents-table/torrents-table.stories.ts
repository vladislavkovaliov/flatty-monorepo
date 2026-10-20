import type { Meta, StoryObj } from '@storybook/vue3-vite'

import TorrentsTable from './torrents-table.vue'

const meta = {
  title: 'shared/ui/TorrentsTable',
  component: TorrentsTable,
  args: {
    items: [
      { title: 'Matrix.1999.1080p.Bluray', mediaType: 'Movie', quality: '1080p', status: "in progress", progress: 30 },
      { title: 'Dune.2021.2160p.WEB-DL', mediaType: 'Movie', quality: '2160p', status: "done", progress: 10 },
      { title: 'Severance.S02.1080p.WEBRip', mediaType: 'Series', quality: '1080p', status: "done", progress: 90 },
    ],
  },
} satisfies Meta<typeof TorrentsTable>

export default meta
type Story = StoryObj<typeof meta>

export const Default: Story = {}
export const Empty: Story = { args: { items: [] } }
