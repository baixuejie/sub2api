<template>
  <div :class="compact ? 'min-w-0' : 'min-w-0 rounded-lg bg-gray-50 px-2.5 py-2 dark:bg-dark-900/60'">
    <div class="truncate text-[10px] uppercase tracking-wide text-gray-400 dark:text-dark-500">{{ label }}</div>
    <div :class="compact ? 'mt-0.5 truncate font-mono text-xs text-gray-600 dark:text-dark-300' : 'mt-0.5 truncate font-mono text-sm font-semibold text-gray-900 dark:text-white'">
      {{ value }}
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  label: string
  value: string
  compact?: boolean
}>()
</script>
