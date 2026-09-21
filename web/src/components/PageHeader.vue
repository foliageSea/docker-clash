<script setup lang="ts">
defineProps<{ title: string; description: string }>()
</script>

<template>
  <header class="mb-6 flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
    <div class="flex min-w-0 flex-col gap-1">
      <h1 class="text-2xl font-semibold">{{ title }}</h1>
      <p class="text-sm text-muted-foreground">{{ description }}</p>
    </div>
    <div v-if="$slots.default" class="flex flex-wrap gap-2"><slot /></div>
  </header>
</template>
