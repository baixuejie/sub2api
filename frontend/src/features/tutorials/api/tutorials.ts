import { apiClient } from '@/api/client'
import type { TutorialVideo } from '../types/tutorials'

export const TUTORIAL_DOCUMENT_URL = 'https://doc.aiprox.net/doc'
export const TUTORIAL_COVER_MAX_BYTES = 5 * 1024 * 1024
const TUTORIAL_VIDEOS_PATH = '/tutorials/videos'

const LOCAL_TUTORIAL_COVER_PATTERN = /^\/api\/v1\/tutorials\/covers\/[a-f0-9]{32}\.(?:png|jpg|webp)$/

export interface TutorialCoverUploadResult {
  filename: string
  url: string
  content_type: 'image/png' | 'image/jpeg' | 'image/webp'
  size: number
}

/** Upload a tutorial cover through the authenticated administrator API. */
export async function uploadTutorialCover(file: File): Promise<TutorialCoverUploadResult> {
  const formData = new FormData()
  formData.append('file', file, file.name || 'tutorial-cover')

  const { data } = await apiClient.post<TutorialCoverUploadResult>(
    '/admin/tutorials/covers',
    formData,
    { headers: { 'Content-Type': undefined } }
  )
  return data
}

/** Accept external HTTP(S) covers and server-generated same-origin cover paths. */
export function isValidTutorialCoverUrl(value: string): boolean {
  const trimmed = value.trim()
  if (!trimmed) return true
  if (LOCAL_TUTORIAL_COVER_PATTERN.test(trimmed)) return true

  try {
    const parsed = new URL(trimmed)
    return parsed.protocol === 'http:' || parsed.protocol === 'https:'
  } catch {
    return false
  }
}

/** Read the public tutorial list without coupling the extension to core settings types. */
export function getPublicTutorialVideos(settings: unknown): TutorialVideo[] {
  if (!settings || typeof settings !== 'object') return []
  const raw = (settings as { tutorial_videos?: unknown }).tutorial_videos
  return normalizePublicTutorialVideos(raw)
}

/** Read the public tutorial list and its server-side play counts. */
export async function fetchPublicTutorialVideos(options?: { signal?: AbortSignal }): Promise<TutorialVideo[]> {
  const { data } = await apiClient.get<unknown>(TUTORIAL_VIDEOS_PATH, {
    signal: options?.signal
  })
  return normalizePublicTutorialVideos(extractVideoList(data))
}

/** Increment one video's play count. The backend performs an atomic increment. */
export async function recordTutorialVideoPlay(
  videoId: string | number,
  options?: { signal?: AbortSignal }
): Promise<number | null> {
  const id = String(videoId).trim()
  if (!id) return null

  const { data } = await apiClient.post<unknown>(
    `${TUTORIAL_VIDEOS_PATH}/${encodeURIComponent(id)}/play`,
    undefined,
    { signal: options?.signal }
  )
  return extractPlayCount(data)
}

function extractVideoList(value: unknown): unknown[] {
  if (Array.isArray(value)) return value
  if (!value || typeof value !== 'object') return []

  const source = value as Record<string, unknown>
  if (Array.isArray(source.videos)) return source.videos
  if (Array.isArray(source.data)) return source.data
  return []
}

function normalizePublicTutorialVideos(raw: unknown): TutorialVideo[] {
  if (!Array.isArray(raw)) return []

  return raw
    .filter((item): item is Record<string, unknown> => Boolean(item) && typeof item === 'object')
    .filter((item) => item.enabled !== false && item.visible !== false)
    .map((item, index) => ({
      id: typeof item.id === 'string' || typeof item.id === 'number' ? item.id : `tutorial-${index}`,
      title: typeof item.title === 'string' ? item.title.trim() : '',
      cover_url: typeof item.cover_url === 'string' ? item.cover_url.trim() : '',
      video_url: typeof item.video_url === 'string' ? item.video_url.trim() : '',
      visible: item.enabled !== false && item.visible !== false,
      sort_order: typeof item.sort_order === 'number' ? item.sort_order : index,
      play_count: normalizePlayCount(item.play_count),
    }))
    .filter((item) => item.title && item.video_url)
    .sort((a, b) => (a.sort_order ?? 0) - (b.sort_order ?? 0))
}

function extractPlayCount(value: unknown): number | null {
  if (typeof value === 'number') return normalizePlayCount(value) ?? null
  if (!value || typeof value !== 'object') return null

  const source = value as Record<string, unknown>
  const direct = source.play_count ?? source.playCount ?? source.count
  if (typeof direct === 'number' || typeof direct === 'string') {
    return normalizePlayCount(direct) ?? null
  }
  return 'data' in source ? extractPlayCount(source.data) : null
}

function normalizePlayCount(value: unknown): number | undefined {
  const count = typeof value === 'number' ? value : typeof value === 'string' ? Number(value) : NaN
  if (!Number.isFinite(count) || count < 0) return undefined
  return Math.trunc(count)
}
