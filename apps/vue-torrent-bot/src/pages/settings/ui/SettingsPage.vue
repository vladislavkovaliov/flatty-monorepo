<script setup lang="ts">
import { useSettings } from '../model/useSettings'

const { data, isPending, isError, error } = useSettings()
</script>

<template>
  <section>
    <h2>Настройки</h2>

    <p v-if="isPending">Загрузка…</p>
    <p v-else-if="isError">Ошибка: {{ error?.message }}</p>

    <ul v-else-if="data">
      <li>Загрузки: {{ data.download_path }}</li>
      <li>Фильмы: {{ data.movies_path }}</li>
      <li>Сериалы: {{ data.tv_path }}</li>
    </ul>
  </section>
</template>
