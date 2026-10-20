import type { Meta, StoryObj } from '@storybook/vue3-vite'

import QualityPicker from './quality-picker.vue'

const meta = {
  title: 'shared/ui/AddTorrentForm/QualityPicker',
  component: QualityPicker,
  args: { modelValue: '1080p' },
} satisfies Meta<typeof QualityPicker>

export default meta
type Story = StoryObj<typeof meta>

export const FullHd: Story = {}
export const UltraHd: Story = { args: { modelValue: '2160p' } }
export const Unknown: Story = { args: { modelValue: 'unknown' } }
