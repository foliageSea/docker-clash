<script setup lang="ts">
defineProps<{ title: string; description?: string }>()
</script>

<template>
  <header class="mb-5 flex flex-wrap items-center justify-between gap-3 sm:mb-6">
    <div class="flex min-w-0 flex-col gap-1">
      <h1 class="text-2xl font-semibold">{{ title }}</h1>
      <p v-if="description" class="text-sm text-muted-foreground">{{ description }}</p>
    </div>
    <div v-if="$slots.default" class="flex flex-wrap gap-2"><slot /></div>
  </header>
</template>
