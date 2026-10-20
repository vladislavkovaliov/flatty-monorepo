<script setup lang="ts">
import { Flex, Text } from '@mantine-vue/core'
import { PhArrowDown } from '@phosphor-icons/vue'

interface Props {
  text?: string
  speed?: string
  showArrow?: boolean
}

withDefaults(defineProps<Props>(), { showArrow: false })
</script>

<template>
  <Flex align="flex-end">
    <Flex>
      <PhArrowDown v-if="showArrow" :size="32" />
      <Text size="xxl" :fw="700">{{ text }}</Text>
    </Flex>
    <Text>{{ speed }}</Text>
  </Flex>
</template>
