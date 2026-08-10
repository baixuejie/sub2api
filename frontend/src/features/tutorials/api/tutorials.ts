import type { TutorialVideo } from '../types/tutorials'

export const TUTORIAL_DOCUMENT_URL = 'https://doc.aiprox.net/doc'

/** Read the public tutorial list without coupling the extension to core settings types. */
export function getPublicTutorialVideos(settings: unknown): TutorialVideo[] {
  if (!settings || typeof settings !== 'object') return []
  const raw = (settings as { tutorial_videos?: unknown }).tutorial_videos
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
    }))
    .filter((item) => item.title && item.video_url)
    .sort((a, b) => (a.sort_order ?? 0) - (b.sort_order ?? 0))
}
