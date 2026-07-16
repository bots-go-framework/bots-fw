package botsfwtest

import "github.com/bots-go-framework/bots-fw/botsfw"

// fakePlatform is a minimal botsfw.BotPlatform for tests.
type fakePlatform struct{}

var _ botsfw.BotPlatform = (*fakePlatform)(nil)

func (p *fakePlatform) ID() string      { return "fake" }
func (p *fakePlatform) Version() string { return "test" }
