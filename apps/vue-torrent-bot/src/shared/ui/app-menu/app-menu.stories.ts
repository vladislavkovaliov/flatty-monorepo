import type { Meta, StoryObj } from '@storybook/vue3-vite'

import AppMenu from './app-menu.vue'

const NAV_ITEMS = [
  { key: 'tasks', label: 'Мои загрузки' },
  { key: 'add', label: 'Добавить' },
  { key: 'settings', label: 'Настройки' },
  { key: 'help', label: 'Помощь' },
]

const meta = {
  title: 'shared/ui/AppMenu',
  component: AppMenu,
  args: {
    items: NAV_ITEMS,
    modelValue: 'tasks',
  },
} satisfies Meta<typeof AppMenu>

export default meta
type Story = StoryObj<typeof meta>

export const Default: Story = {}

export const FakeItems: Story = {
  args: {
    items: [
      { key: 'first', label: 'Первый пункт' },
      { key: 'second', label: 'Второй пункт' },
      { key: 'third', label: 'Третий пункт' },
    ],
    modelValue: 'second',
  },
}
