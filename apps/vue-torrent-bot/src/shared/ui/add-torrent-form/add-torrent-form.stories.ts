import type { Meta, StoryObj } from '@storybook/vue3-vite'

import AddTorrentForm from './add-torrent-form.vue'

const FULL_MAGNET =
  'magnet:?xt=urn:btih:A1B2C3D4E5F60718293A4B5C6D7E8F9012345678&dn=The.Matrix.1999.1080p.BluRay&tr=udp://tracker.opentrackr.org:1337&tr=udp://open.tracker.cl:1337&xl=2147483648'

const meta = {
  title: 'shared/ui/AddTorrentForm',
  component: AddTorrentForm,
  args: { initialMagnet: FULL_MAGNET },
} satisfies Meta<typeof AddTorrentForm>

export default meta
type Story = StoryObj<typeof meta>

export const Empty: Story = { args: { initialMagnet: '' } }
export const Details: Story = {}
export const Confirm: Story = { args: { initialStep: 'confirm' } }
export const InvalidMagnet: Story = { args: { initialMagnet: 'https://example.com/file.torrent' } }
export const ServerError: Story = {
  args: { initialStep: 'confirm', error: 'Такая загрузка уже есть' },
}
export const Submitting: Story = { args: { initialStep: 'confirm', submitting: true } }
export const ProductionMode: Story = { args: { initialStep: 'confirm', testMode: false } }
