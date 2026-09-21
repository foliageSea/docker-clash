<script setup lang="ts">
import { onMounted } from 'vue'

import AppShell from '@/components/AppShell.vue'
import { Toaster } from '@/components/ui/sonner'
import { useConsole } from '@/composables/use-console'

const { load } = useConsole()
onMounted(load)
</script>

<template>
  <Toaster position="top-right" rich-colors close-button />
  <AppShell />
</template>
