import type { Meta, StoryObj } from '@storybook/vue3-vite'

import CategoryPicker from './category-picker.vue'

const meta = {
  title: 'shared/ui/AddTorrentForm/CategoryPicker',
  component: CategoryPicker,
  args: { modelValue: 'movie' },
} satisfies Meta<typeof CategoryPicker>

export default meta
type Story = StoryObj<typeof meta>

export const Movie: Story = {}
export const Series: Story = { args: { modelValue: 'series' } }
