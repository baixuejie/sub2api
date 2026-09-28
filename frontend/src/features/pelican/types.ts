export interface GroupTag {
  id: number
  name: string
}
export interface PelicanConfig {
  revision: number
  enabled: boolean
  selected_group_ids: number[]
  topic_mode: 'rotate' | 'fixed'
  fixed_topic_id: string
  max_output_tokens: number
  timeout_seconds: number
  retention_days: number
  interval_minutes: 10 | 30 | 60
  next_run_at: string | null
  key_configured: boolean
  key_masked: string
  encryption_ready: boolean
  key_unavailable: boolean
  model: string
  reasoning_effort: string
}
export type SaveConfig = Pick<
  PelicanConfig,
  | 'revision'
  | 'enabled'
  | 'selected_group_ids'
  | 'topic_mode'
  | 'fixed_topic_id'
  | 'max_output_tokens'
  | 'timeout_seconds'
  | 'retention_days'
  | 'interval_minutes'
> & { api_key?: string }
export interface PelicanRun {
  id: number
  scheduled_for: string
  finished_at: string | null
  status: string
  topic_id: string
  groups: GroupTag[]
  model: string
  reasoning_effort: string
  input_tokens: number | null
  output_tokens: number | null
  total_tokens: number | null
  latency_ms: number | null
  error_code?: string
}
export interface RunPage {
  items: PelicanRun[]
  next_cursor: string
}
export interface GalleryStatus {
  enabled: boolean
  next_run_at: string | null
  last_success_at: string | null
  interval_seconds: number
}
export interface RunDetail {
  run: PelicanRun
  prompt: string
  request: unknown
  usage: unknown
  skipped_hours: number
  source_available: boolean
  raw_path: string
  preview_path: string
  preview_notes: string[]
}
export const topicIDs = ['pelican-ski', 'wukong-airplane', 'polar-bear-ultraman'] as const
export function formatTime(value: string | null, locale: string): string {
  return value
    ? new Date(value).toLocaleString(locale === 'zh' ? 'zh-CN' : 'en-US', { hour12: false })
    : '—'
}
