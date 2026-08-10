import { describe, expect, it } from 'vitest'
import { getPublicTutorialVideos, TUTORIAL_DOCUMENT_URL } from '../api/tutorials'

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
})
