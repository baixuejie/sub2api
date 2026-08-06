<template>
  <aside class="hidden lg:block">
    <div class="sticky top-24">
      <div class="mb-3 flex items-center justify-between gap-2 px-1">
        <h2 class="text-xs font-semibold uppercase tracking-[0.16em] text-gray-400 dark:text-dark-500">
          {{ copy.groupNavigation }}
        </h2>
        <span class="font-mono text-xs text-gray-400 dark:text-dark-500">{{ groups.length }}</span>
      </div>
      <nav class="space-y-1.5" :aria-label="copy.groupNavigation">
        <button
          v-for="group in groups"
          :key="group.id"
          type="button"
          class="group-nav-item"
          :class="{ 'group-nav-item-active': group.id === selectedId }"
          :style="group.id === selectedId ? { '--group-accent': platformAccentColor(group.platform) } : undefined"
          @click="$emit('select', group.id)"
        >
          <span class="h-8 w-1 shrink-0 rounded-full bg-gray-200 transition-colors dark:bg-dark-700" :style="{ backgroundColor: group.id === selectedId ? platformAccentColor(group.platform) : undefined }"></span>
          <span class="min-w-0 flex-1 text-left">
            <span class="flex items-center gap-1.5 truncate text-sm font-medium text-gray-800 dark:text-dark-100">
              <PlatformIcon :platform="group.platform as GroupPlatform" size="xs" :style="{ color: platformAccentColor(group.platform) }" />
              <span class="truncate">{{ group.name }}</span>
            </span>
            <span class="mt-0.5 block text-xs text-gray-400 dark:text-dark-500">
              {{ group.models.length }} {{ copy.modelCount }}
            </span>
          </span>
          <Icon name="chevronRight" size="xs" class="shrink-0 text-gray-300 transition-transform group-hover:translate-x-0.5 dark:text-dark-600" aria-hidden="true" />
        </button>
      </nav>
    </div>
  </aside>
</template>

<script setup lang="ts">
import Icon from '@/components/icons/Icon.vue'
import PlatformIcon from '@/components/common/PlatformIcon.vue'
import type { GroupPlatform } from '@/types'
import { platformAccentColor } from '@/utils/platformColors'
import { useModelPlazaLocale } from '../locales'
import type { ModelPlazaGroup } from '../types/modelPlaza'

defineProps<{
  groups: ModelPlazaGroup[]
  selectedId: number | null
}>()

defineEmits<{
  select: [id: number]
}>()

const copy = useModelPlazaLocale()
</script>

<style scoped>
.group-nav-item {
  @apply flex w-full items-center gap-3 rounded-xl px-2.5 py-2.5 text-left transition-colors;
  @apply hover:bg-gray-100 dark:hover:bg-dark-800;
}

.group-nav-item-active {
  @apply bg-white shadow-sm ring-1 ring-gray-200 dark:bg-dark-800/80 dark:ring-dark-700;
}
</style>
