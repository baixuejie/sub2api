package tutorials

import "context"

// PublicVideosProvider reads the currently enabled and validated tutorial
// videos from the core settings service.
//
// The tutorials extension owns the HTTP contract while the server assembly
// supplies this small adapter, keeping the extension independent from the
// concrete SettingService implementation.
type PublicVideosProvider func(context.Context) ([]Video, error)

// PlayCountStore persists tutorial video play counts.
type PlayCountStore interface {
	List(context.Context, []string) (map[string]int64, error)
	Increment(context.Context, string) (int64, error)
}
