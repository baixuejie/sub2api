import { describe, expect, it } from 'vitest'
import {
  getPublicTutorialVideos,
  isValidTutorialCoverUrl,
  TUTORIAL_DOCUMENT_URL,
} from '../api/tutorials'

describe('tutorials extension', () => {
  it('reads visible videos, filters incomplete entries, and sorts by sort_order', () => {
    const videos = getPublicTutorialVideos({
      tutorial_videos: [
        { id: 'hidden', title: 'Hidden', video_url: 'https://example.com/h', visible: false, sort_order: 0 },
        { id: 'second', title: 'Second', video_url: 'https://example.com/2', sort_order: 20 },
        { id: 'first', title: 'First', cover_url: 'https://example.com/1.jpg', video_url: 'https://example.com/1', sort_order: 10 },
        { id: 'invalid', title: '', video_url: 'https://example.com/i' },
      ],
    })

    expect(videos.map((video) => video.id)).toEqual(['first', 'second'])
  })

  it('uses the fixed documentation URL and handles malformed settings', () => {
    expect(TUTORIAL_DOCUMENT_URL).toBe('https://doc.aiprox.net/doc')
    expect(getPublicTutorialVideos(null)).toEqual([])
    expect(getPublicTutorialVideos({ tutorial_videos: 'invalid' })).toEqual([])
  })

  it('accepts uploaded local covers without opening arbitrary relative paths', () => {
    expect(isValidTutorialCoverUrl('/api/v1/tutorials/covers/0123456789abcdef0123456789abcdef.webp')).toBe(true)
    expect(isValidTutorialCoverUrl('https://example.com/cover.png')).toBe(true)
    expect(isValidTutorialCoverUrl('/api/v1/tutorials/covers/../config.yaml')).toBe(false)
    expect(isValidTutorialCoverUrl('/api/v1/settings')).toBe(false)
  })
})
