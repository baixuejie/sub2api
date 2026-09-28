package pelican

import (
	"crypto/sha256"
	"embed"
	"fmt"
	"strings"
)

//go:embed prompts/*.txt
var promptFiles embed.FS

type Topic struct {
	ID      string `json:"id"`
	Version string `json:"version"`
}

var topics = []Topic{{"pelican-ski", "v1"}, {"wukong-airplane", "v1"}, {"polar-bear-ultraman", "v1"}}

func topicByID(id string) *Topic {
	for _, t := range topics {
		if t.ID == id {
			return &t
		}
	}
	return nil
}

func selectTopic(cfg Config, skipped int64) Topic {
	if cfg.TopicMode == "fixed" {
		if t := topicByID(cfg.FixedTopicID); t != nil {
			return *t
		}
	}
	return topics[(cfg.Sequence+skipped)%int64(len(topics))]
}

func (t Topic) prompt() string {
	scene, _ := promptFiles.ReadFile("prompts/" + t.ID + ".txt")
	common, _ := promptFiles.ReadFile("prompts/common.txt")
	return strings.TrimSpace(string(scene)) + strings.TrimSpace(string(common))
}

func contentHash(s string) string { return fmt.Sprintf("%x", sha256.Sum256([]byte(s))) }
