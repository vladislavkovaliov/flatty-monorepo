import { Text } from '@mantine-vue/core'
import type { Meta, StoryObj } from '@storybook/vue3-vite'

import HeroCard from './hero-card.vue'

const meta = {
  title: 'shared/ui/HeroCard',
  component: HeroCard,
  argTypes: {
    type: { control: 'select', options: ['download', 'done', 'speed'] },
    text: { control: 'text' },
  },
} satisfies Meta<typeof HeroCard>

export default meta
type Story = StoryObj<typeof meta>

// The #header slot content lives in markup, so each story sets it while props come from args.
const headerStory = (header: string, args: NonNullable<Story['args']>): Story => ({
  args,
  render: (storyArgs) => ({
    components: { HeroCard, Text },
    setup: () => ({ storyArgs }),
    template: `<HeroCard v-bind="storyArgs"><template #header><Text c="black">${header}</Text></template></HeroCard>`,
  }),
})

export const Active: Story = headerStory('Active', { type: 'download', text: '12' })
export const Done: Story = headerStory('Done', { type: 'done', text: '48' })
export const Speed: Story = headerStory('Speed', { type: 'speed', text: '12' })
