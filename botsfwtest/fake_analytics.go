package botsfwtest

import "github.com/strongo/analytics"

// fakeAnalytics implements botsfw.WebhookAnalytics without any network calls.
// Enqueued messages are silently dropped; callers that need to inspect analytics
// events should substitute their own implementation via the option.
type fakeAnalytics struct {
	messages []analytics.Message
}

// Enqueue records the message in memory.
func (a *fakeAnalytics) Enqueue(msg analytics.Message) {
	a.messages = append(a.messages, msg)
}

// Messages returns all enqueued analytics messages.
func (a *fakeAnalytics) Messages() []analytics.Message {
	return a.messages
}
