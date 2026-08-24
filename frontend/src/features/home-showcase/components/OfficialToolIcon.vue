<template>
  <img
    v-if="localIcon"
    class="official-tool-icon official-tool-icon-image"
    :src="localIcon"
    :alt="toolLabel"
    :width="size"
    :height="size"
    loading="lazy"
    decoding="async"
  />
  <ModelIcon
    v-else
    class="official-tool-icon official-tool-icon-mark"
    :model="brandModel"
    :size="`${size}px`"
    :aria-label="toolLabel"
  />
</template>

<script setup lang="ts">
import { computed } from 'vue'
import ModelIcon from '@/components/common/ModelIcon.vue'
import openclawIcon from '../assets/openclaw-official.svg'
import ccswitchIcon from '../assets/ccswitch-official.png'
import deepseekHarnessIcon from '../assets/deepseek-harness-official.svg'
import hermesIcon from '../assets/hermes-official.png'

defineOptions({ name: 'OfficialToolIcon' })

type ToolId = 'claude' | 'codex' | 'openclaw' | 'hermes' | 'deepseek-harness' | 'ccswitch'

const props = withDefaults(defineProps<{
  tool: ToolId
  size?: number
}>(), {
  size: 26
})

const localIcons: Partial<Record<ToolId, string>> = {
  // Official assets bundled locally so the homepage remains usable offline.
  openclaw: openclawIcon,
  ccswitch: ccswitchIcon,
  hermes: hermesIcon,
  'deepseek-harness': deepseekHarnessIcon
}

const toolLabel = computed(() => ({
  claude: 'Claude',
  codex: 'Codex',
  openclaw: 'OpenClaw',
  hermes: 'Hermes',
  'deepseek-harness': 'DeepSeek Harness',
  ccswitch: 'CC Switch'
}[props.tool]))

const brandModel = computed(() => {
  if (props.tool === 'claude') return 'claude-3-7-sonnet'
  if (props.tool === 'codex') return 'codex'
  if (props.tool === 'deepseek-harness') return 'deepseek'
  return 'openai'
})

const localIcon = computed(() => localIcons[props.tool])
</script>

<style scoped>
.official-tool-icon {
  display: block;
  flex: 0 0 auto;
  max-width: 100%;
  max-height: 100%;
}

.official-tool-icon-image {
  object-fit: contain;
}

.official-tool-icon-mark {
  overflow: hidden;
}

</style>
