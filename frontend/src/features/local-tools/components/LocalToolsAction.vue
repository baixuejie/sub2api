<template>
  <div v-if="tools.length > 0" ref="rootRef" class="relative inline-flex">
    <button
      ref="triggerRef"
      type="button"
      class="flex flex-col items-center gap-0.5 rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-primary-50 hover:text-primary-600 disabled:cursor-not-allowed disabled:opacity-40 dark:hover:bg-primary-900/20 dark:hover:text-primary-400"
      :disabled="allToolsDisabled"
      :title="allToolsDisabled ? copy.unavailable : copy.menuTitle"
      aria-haspopup="menu"
      :aria-controls="menuOpen ? menuId : undefined"
      :aria-expanded="menuOpen"
      @click.stop="toggleMenu"
      @keydown.down.prevent="openMenu('first')"
      @keydown.up.prevent="openMenu('last')"
    >
      <span class="flex items-center gap-0.5">
        <Icon name="terminal" size="sm" />
        <Icon name="chevronDown" size="xs" class="opacity-70" />
      </span>
      <span class="text-xs">{{ copy.menuLabel }}</span>
    </button>

    <component
      :is="tool.actionComponent"
      v-for="tool in tools"
      :key="`action-${tool.id}`"
      :ref="(instance: unknown) => setActionRef(tool.id, instance)"
      v-bind="tool.actionProps(toolContext)"
    >
      <template #trigger><span class="hidden" aria-hidden="true"></span></template>
    </component>

    <Teleport to="body">
      <div
        v-if="menuOpen && menuPosition"
        :id="menuId"
        ref="menuRef"
        class="fixed z-[100000030] w-64 max-w-[calc(100vw-16px)] overflow-hidden rounded-xl border border-gray-200 bg-white p-1.5 shadow-xl shadow-gray-900/10 dark:border-dark-600 dark:bg-dark-800 dark:shadow-black/30"
        :style="{ top: `${menuPosition.top}px`, left: `${menuPosition.left}px` }"
        role="menu"
        :aria-label="copy.menuTitle"
        @click.stop
        @keydown="handleMenuKeydown"
      >
        <div class="px-2.5 pb-1.5 pt-1 text-[11px] font-semibold uppercase tracking-wide text-gray-400 dark:text-dark-500">
          {{ copy.selectTool }}
        </div>
        <button
          v-for="tool in tools"
          :key="tool.id"
          type="button"
          class="flex w-full items-center gap-3 rounded-lg px-2.5 py-2 text-left transition-colors disabled:cursor-not-allowed disabled:opacity-50"
          :class="tool.hoverClass"
          role="menuitem"
          tabindex="-1"
          :data-tool-id="tool.id"
          :disabled="tool.isDisabled(toolContext)"
          :title="tool.isDisabled(toolContext) ? copy.unavailable : copy[tool.descriptionKey]"
          @click="selectTool(tool)"
        >
          <span
            class="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg"
            :class="tool.iconClass"
          >
            <Icon :name="tool.icon" size="sm" />
          </span>
          <span class="min-w-0">
            <span class="block text-sm font-medium text-gray-900 dark:text-white">
              {{ copy[tool.labelKey] }}
            </span>
            <span class="block truncate text-xs text-gray-500 dark:text-dark-400">
              {{ copy[tool.descriptionKey] }}
            </span>
          </span>
        </button>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import type { ApiKey, PublicSettings } from '@/types'
import Icon from '@/components/icons/Icon.vue'
import {
  resolveLocalTools,
  type LocalToolDefinition
} from '../localToolRegistry'
import { localToolsCopy } from '../localToolsCopy'

interface Props {
  apiKey: ApiKey
  publicSettings: PublicSettings | null
}

interface LocalToolActionHandle {
  openDialog: () => void
}

type InitialFocus = 'first' | 'last'

const props = defineProps<Props>()
const { locale } = useI18n()
const copy = computed(() => localToolsCopy[locale.value.toLowerCase().startsWith('zh') ? 'zh' : 'en'])
const toolContext = computed(() => ({
  apiKey: props.apiKey,
  publicSettings: props.publicSettings
}))
const tools = computed(() => resolveLocalTools(toolContext.value))
const allToolsDisabled = computed(() =>
  tools.value.length > 0 && tools.value.every((tool) => tool.isDisabled(toolContext.value))
)

const rootRef = ref<HTMLElement | null>(null)
const triggerRef = ref<HTMLButtonElement | null>(null)
const menuRef = ref<HTMLElement | null>(null)
const menuOpen = ref(false)
const menuPosition = ref<{ top: number; left: number } | null>(null)
const actionRefs = new Map<string, LocalToolActionHandle>()
const menuId = `local-tools-menu-${Math.random().toString(36).slice(2, 10)}`

function setActionRef(toolId: string, instance: unknown): void {
  if (instance && typeof (instance as LocalToolActionHandle).openDialog === 'function') {
    actionRefs.set(toolId, instance as LocalToolActionHandle)
  } else {
    actionRefs.delete(toolId)
  }
}

function updateMenuPosition(): void {
  const trigger = triggerRef.value
  if (!trigger) return
  const rect = trigger.getBoundingClientRect()
  const width = Math.min(256, window.innerWidth - 16)
  const left = Math.max(8, Math.min(rect.right - width, window.innerWidth - width - 8))
  const estimatedHeight = 42 + tools.value.length * 56
  const top = window.innerHeight - rect.bottom >= estimatedHeight + 8
    ? rect.bottom + 6
    : Math.max(8, rect.top - estimatedHeight - 6)
  menuPosition.value = { top, left }
}

async function openMenu(initialFocus: InitialFocus = 'first'): Promise<void> {
  if (allToolsDisabled.value || menuOpen.value) return
  menuOpen.value = true
  await nextTick()
  updateMenuPosition()
  addGlobalListeners()
  await nextTick()
  focusBoundaryItem(initialFocus)
}

async function toggleMenu(): Promise<void> {
  if (menuOpen.value) {
    closeMenu()
    return
  }
  await openMenu()
}

function closeMenu(restoreFocus = false): void {
  menuOpen.value = false
  menuPosition.value = null
  removeGlobalListeners()
  if (restoreFocus) void nextTick(() => triggerRef.value?.focus())
}

function selectTool(tool: LocalToolDefinition): void {
  if (tool.isDisabled(toolContext.value)) return
  closeMenu()
  actionRefs.get(tool.id)?.openDialog()
}

function enabledMenuItems(): HTMLButtonElement[] {
  if (!menuRef.value) return []
  return [...menuRef.value.querySelectorAll<HTMLButtonElement>('[role="menuitem"]:not(:disabled)')]
}

function focusBoundaryItem(position: InitialFocus): void {
  const items = enabledMenuItems()
  if (items.length === 0) return
  items[position === 'first' ? 0 : items.length - 1]?.focus()
}

function moveMenuFocus(offset: number): void {
  const items = enabledMenuItems()
  if (items.length === 0) return
  const currentIndex = items.indexOf(document.activeElement as HTMLButtonElement)
  const nextIndex = currentIndex === -1
    ? (offset > 0 ? 0 : items.length - 1)
    : (currentIndex + offset + items.length) % items.length
  items[nextIndex]?.focus()
}

function handleMenuKeydown(event: KeyboardEvent): void {
  switch (event.key) {
    case 'ArrowDown':
      event.preventDefault()
      moveMenuFocus(1)
      break
    case 'ArrowUp':
      event.preventDefault()
      moveMenuFocus(-1)
      break
    case 'Home':
      event.preventDefault()
      focusBoundaryItem('first')
      break
    case 'End':
      event.preventDefault()
      focusBoundaryItem('last')
      break
    case 'Escape':
      event.preventDefault()
      closeMenu(true)
      break
  }
}

function handleOutsidePointerDown(event: PointerEvent): void {
  const target = event.target as Node | null
  if (target && (rootRef.value?.contains(target) || menuRef.value?.contains(target))) return
  closeMenu()
}

function addGlobalListeners(): void {
  window.addEventListener('resize', updateMenuPosition)
  window.addEventListener('scroll', updateMenuPosition, true)
  document.addEventListener('pointerdown', handleOutsidePointerDown, true)
}

function removeGlobalListeners(): void {
  window.removeEventListener('resize', updateMenuPosition)
  window.removeEventListener('scroll', updateMenuPosition, true)
  document.removeEventListener('pointerdown', handleOutsidePointerDown, true)
}

watch(() => tools.value.length, (length) => {
  if (length === 0 && menuOpen.value) closeMenu()
})

onBeforeUnmount(() => {
  removeGlobalListeners()
  actionRefs.clear()
})
</script>
