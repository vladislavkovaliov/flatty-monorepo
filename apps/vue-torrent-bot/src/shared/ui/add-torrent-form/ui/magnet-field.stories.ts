import type { Meta, StoryObj } from '@storybook/vue3-vite'

import type { MagnetInfo } from '@/entities/task'

import MagnetField from './magnet-field.vue'

const FULL_MAGNET =
  'magnet:?xt=urn:btih:A1B2C3D4E5F60718293A4B5C6D7E8F9012345678&dn=The.Matrix.1999.1080p.BluRay&tr=udp://tracker.opentrackr.org:1337&tr=udp://open.tracker.cl:1337&xl=2147483648'

const PARSED: MagnetInfo = {
  raw: FULL_MAGNET,
  infoHash: 'A1B2C3D4E5F60718293A4B5C6D7E8F9012345678',
  displayName: 'The.Matrix.1999.1080p.BluRay',
  trackers: ['udp://tracker.opentrackr.org:1337', 'udp://open.tracker.cl:1337'],
  sizeBytes: 2147483648,
}

const meta = {
  title: 'shared/ui/AddTorrentForm/MagnetField',
  component: MagnetField,
  args: { modelValue: FULL_MAGNET, parsed: PARSED },
} satisfies Meta<typeof MagnetField>

export default meta
type Story = StoryObj<typeof meta>

export const Empty: Story = { args: { modelValue: '', parsed: null } }
export const Parsed: Story = {}
export const Invalid: Story = {
  args: {
    modelValue: 'https://example.com/file.torrent',
    parsed: null,
    error: 'Ссылка не распознана',
  },
}
export const Disabled: Story = { args: { disabled: true } }
