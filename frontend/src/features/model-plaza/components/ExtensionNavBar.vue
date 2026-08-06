<template>
  <header class="glass sticky top-0 z-30 border-b border-gray-200/60 dark:border-dark-700/60">
    <div class="mx-auto flex max-w-7xl items-center justify-between gap-4 px-4 py-3 sm:px-6">
      <RouterLink to="/model-plaza" class="flex min-w-0 items-center gap-3" aria-label="Model Plaza">
        <span
          v-if="settings"
          class="flex h-9 w-9 shrink-0 items-center justify-center overflow-hidden rounded-xl bg-white shadow-sm ring-1 ring-gray-200 dark:bg-dark-800 dark:ring-dark-700"
        >
          <img :src="siteLogo || '/logo.svg'" :alt="siteName" class="h-full w-full object-contain" />
        </span>
        <span v-else class="h-9 w-9 shrink-0 animate-pulse rounded-xl bg-gray-200 dark:bg-dark-700" aria-hidden="true"></span>
        <span class="truncate text-base font-semibold text-gray-950 dark:text-white">{{ siteName }}</span>
      </RouterLink>

      <RouterLink
        v-if="isAuthenticated"
        :to="backTarget"
        class="btn btn-primary btn-sm shrink-0"
      >
        <Icon name="arrowLeft" size="xs" aria-hidden="true" />
        <span>{{ t('modelPlaza.nav.backToDashboard') }}</span>
      </RouterLink>
      <RouterLink
        v-else
        :to="{ path: '/login', query: { redirect: '/model-plaza' } }"
        class="btn btn-primary btn-sm shrink-0"
      >
        <Icon name="login" size="xs" aria-hidden="true" />
        <span>{{ t('modelPlaza.nav.login') }}</span>
      </RouterLink>
    </div>
  </header>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import { sanitizeUrl } from '@/utils/url'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'

const { t } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()

const settings = computed(() => appStore.cachedPublicSettings)
const siteName = computed(() => settings.value?.site_name || 'Sub2API')
const siteLogo = computed(() =>
  sanitizeUrl(settings.value?.site_logo || '', { allowRelative: true, allowDataUrl: true })
)
const isAuthenticated = computed(() => authStore.isAuthenticated)
const backTarget = computed(() => (authStore.isAdmin ? '/admin/dashboard' : '/dashboard'))
</script>
