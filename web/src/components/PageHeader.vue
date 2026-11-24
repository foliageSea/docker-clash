<script setup lang="ts">
defineProps<{ title: string }>()
</script>

<template>
  <header class="mb-5 flex flex-wrap items-end justify-between gap-3 md:mb-7 md:gap-4">
    <h1 class="min-w-0 text-xl font-semibold tracking-tight sm:text-2xl">{{ title }}</h1>
    <div v-if="$slots.default" class="flex flex-wrap gap-2"><slot /></div>
  </header>
</template>
