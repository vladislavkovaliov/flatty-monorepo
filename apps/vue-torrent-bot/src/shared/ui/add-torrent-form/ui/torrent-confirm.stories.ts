import type { Meta, StoryObj } from '@storybook/vue3-vite'

import TorrentConfirm from './torrent-confirm.vue'

const meta = {
  title: 'shared/ui/AddTorrentForm/TorrentConfirm',
  component: TorrentConfirm,
  args: { name: 'The.Matrix.1999.1080p.BluRay', mediaType: 'movie', quality: '1080p' },
} satisfies Meta<typeof TorrentConfirm>

export default meta
type Story = StoryObj<typeof meta>

export const TestMode: Story = {}
export const ProductionMode: Story = { args: { testMode: false } }
export const Series: Story = { args: { mediaType: 'series', quality: '2160p' } }
export const Submitting: Story = { args: { submitting: true } }
