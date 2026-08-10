package service

import tutorials "github.com/Wei-Shaw/sub2api/internal/extensions/tutorials"

const TutorialDocumentURL = tutorials.DocumentURL

type TutorialVideo = tutorials.Video

func NormalizeTutorialVideosJSON(raw string) (string, error) {
	return tutorials.NormalizeVideosJSON(raw)
}

func ParsePublicTutorialVideos(raw string) []TutorialVideo {
	return tutorials.ParsePublicVideos(raw)
}
