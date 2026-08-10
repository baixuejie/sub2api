<template>
  <div class="card">
    <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
      <div class="flex items-start justify-between gap-4">
        <div>
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
            {{ localText('使用教程视频', 'Tutorial videos') }}
          </h2>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
            {{
              localText(
                '配置后会在使用教程页面展示已启用的视频，用户点击后在新窗口播放。文档固定嵌入 https://doc.aiprox.net/doc。',
                'Configure video guides shown on the Tutorials page. Enabled videos open in a new window. Documentation is embedded from https://doc.aiprox.net/doc.'
              )
            }}
          </p>
        </div>
        <Icon name="book" size="md" class="flex-shrink-0 text-primary-500" aria-hidden="true" />
      </div>
    </div>

    <div class="p-6">
      <div
        v-if="videos.length"
        class="divide-y divide-gray-200 rounded-lg border border-gray-200 dark:divide-dark-700 dark:border-dark-600"
      >
        <div
          v-for="(video, index) in videos"
          :key="video.id || index"
          class="space-y-4 p-4"
        >
          <div class="flex items-center justify-between gap-3">
            <div class="flex min-w-0 items-center gap-2">
              <span class="text-sm font-semibold text-gray-900 dark:text-white">
                {{ localText('视频教程', 'Video') }} {{ index + 1 }}
              </span>
              <span
                class="rounded-full px-2 py-0.5 text-xs"
                :class="
                  video.enabled
                    ? 'bg-emerald-50 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300'
                    : 'bg-gray-100 text-gray-500 dark:bg-dark-700 dark:text-gray-400'
                "
              >
                {{ video.enabled ? localText('展示中', 'Visible') : localText('已隐藏', 'Hidden') }}
              </span>
            </div>
            <div class="flex flex-shrink-0 items-center gap-1">
              <button
                v-if="index > 0"
                type="button"
                class="rounded p-1.5 text-gray-400 hover:bg-gray-100 hover:text-gray-700 dark:hover:bg-dark-700 dark:hover:text-gray-200"
                :title="localText('上移', 'Move up')"
                @click="moveVideo(index, -1)"
              >
                <Icon name="arrowUp" size="sm" aria-hidden="true" />
              </button>
              <button
                v-if="index < videos.length - 1"
                type="button"
                class="rounded p-1.5 text-gray-400 hover:bg-gray-100 hover:text-gray-700 dark:hover:bg-dark-700 dark:hover:text-gray-200"
                :title="localText('下移', 'Move down')"
                @click="moveVideo(index, 1)"
              >
                <Icon name="arrowDown" size="sm" aria-hidden="true" />
              </button>
              <button
                type="button"
                class="rounded p-1.5 text-red-400 hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-900/20"
                :title="localText('删除视频', 'Remove video')"
                @click="removeVideo(index)"
              >
                <Icon name="trash" size="sm" aria-hidden="true" />
              </button>
            </div>
          </div>

          <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
            <div>
              <label class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-gray-300">
                {{ localText('标题', 'Title') }}
              </label>
              <input
                v-model="video.title"
                type="text"
                maxlength="200"
                class="input"
                :placeholder="localText('例如：API Key 创建与调用', 'Example: Create and use an API key')"
              />
            </div>
            <div>
              <label class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-gray-300">
                {{ localText('视频链接', 'Video URL') }}
              </label>
              <input v-model="video.video_url" type="url" class="input font-mono text-sm" placeholder="https://" />
            </div>
            <div class="md:col-span-2">
              <label class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-gray-300">
                {{ localText('封面', 'Cover') }}
              </label>
              <input
                v-model="video.cover_url"
                type="url"
                class="input font-mono text-sm"
                placeholder="https://.../cover.jpg"
              />
              <p class="mt-1.5 text-xs text-gray-500 dark:text-gray-400">
                {{ localText('可选，建议使用稳定的 HTTPS 图片地址。', 'Optional. Use a stable HTTPS image URL.') }}
              </p>
            </div>
            <div
              class="md:col-span-2 flex items-center justify-between border-t border-gray-100 pt-3 dark:border-dark-700"
            >
              <div>
                <label class="font-medium text-gray-900 dark:text-white">
                  {{ localText('是否展示', 'Show video') }}
                </label>
                <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
                  {{
                    localText(
                      '关闭后仅保存在后台，不会出现在用户教程页面。',
                      'When disabled, the video remains saved but is hidden from users.'
                    )
                  }}
                </p>
              </div>
              <Toggle v-model="video.enabled" />
            </div>
          </div>
        </div>
      </div>

      <div
        v-else
        class="rounded-lg border border-dashed border-gray-300 px-4 py-6 text-center text-sm text-gray-500 dark:border-dark-600 dark:text-gray-400"
      >
        {{ localText('还没有配置视频教程。', 'No video tutorials configured yet.') }}
      </div>

      <button
        type="button"
        class="mt-4 flex w-full items-center justify-center gap-2 rounded-lg border-2 border-dashed border-gray-300 py-3 text-sm text-gray-500 transition-colors hover:border-primary-400 hover:text-primary-600 dark:border-dark-600 dark:text-gray-400 dark:hover:border-primary-500 dark:hover:text-primary-400"
        @click="addVideo"
      >
        <Icon name="plus" size="sm" aria-hidden="true" />
        {{ localText('添加视频教程', 'Add video tutorial') }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { TutorialVideoSetting } from '@/api/admin/settings'
import Icon from '@/components/icons/Icon.vue'
import Toggle from '@/components/common/Toggle.vue'

const videos = defineModel<TutorialVideoSetting[]>({ required: true })
const { locale } = useI18n()
const isZhLocale = computed(() => locale.value.startsWith('zh'))

function localText(zh: string, en: string): string {
  return isZhLocale.value ? zh : en
}

function normalizeSortOrder(): void {
  videos.value.forEach((video, index) => {
    video.sort_order = index
  })
}

function addVideo(): void {
  videos.value.push({
    id: `tutorial-${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 8)}`,
    title: '',
    cover_url: '',
    video_url: '',
    enabled: true,
    sort_order: videos.value.length
  })
}

function removeVideo(index: number): void {
  videos.value.splice(index, 1)
  normalizeSortOrder()
}

function moveVideo(index: number, direction: -1 | 1): void {
  const targetIndex = index + direction
  if (targetIndex < 0 || targetIndex >= videos.value.length) return
  const current = videos.value[index]
  videos.value[index] = videos.value[targetIndex]
  videos.value[targetIndex] = current
  normalizeSortOrder()
}
</script>
