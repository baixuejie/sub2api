export interface TutorialVideo {
  id: string | number
  title: string
  cover_url: string
  video_url: string
  visible?: boolean
  sort_order?: number
  play_count?: number
}
