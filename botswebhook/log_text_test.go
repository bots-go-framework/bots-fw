package botswebhook

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/bots-go-framework/bots-fw/botinput"
	"github.com/bots-go-framework/bots-fw/botmsg"
	"github.com/bots-go-framework/bots-fw/botsfw"
	"github.com/strongo/logus"
	"go.uber.org/mock/gomock"
)

func TestBotLogText(t *testing.T) {
	for _, text := range []string{"/start watch-local_nonce", "/start@DataTugBot watch-local_nonce", " \t/start@DataTugBot\nwatch-local_nonce extra-secret", "/START@AnotherBot   watch-local_nonce"} {
		if got := botLogText(text); strings.Contains(got, "nonce") || strings.Contains(got, "extra-secret") || !strings.Contains(got, "[redacted]") {
			t.Fatalf("bearer exposed in %q", got)
		}
	}
	for _, text := range []string{"", "/start", "/start ordinary", "/starting watch-user-text", "ordinary watch-user-text"} {
		if got := botLogText(text); got != text {
			t.Errorf("ordinary text changed %q=>%q", text, got)
		}
	}
}

type watchLogCaptureKey struct{}
type watchLogCapture struct {
	mu       sync.Mutex
	messages []string
}

func (c *watchLogCapture) Log(ctx context.Context, entry logus.LogEntry) error {
	if ctx.Value(watchLogCaptureKey{}) == c {
		c.mu.Lock()
		c.messages = append(c.messages, fmt.Sprintf(entry.MessageFormat, entry.MessageArgs...))
		c.mu.Unlock()
	}
	return nil
}
func TestWatchNonceLogsRedactedExecutionUnchanged(t *testing.T) {
	capture := &watchLogCapture{}
	logus.AddLogEntryHandler(capture)
	ctx := context.WithValue(context.Background(), watchLogCaptureKey{}, capture)
	original := "/start@DataTugBot watch-local_private-nonce"
	ctrl := gomock.NewController(t)
	input := &logInputText{logInputBase: newLogInputBase(ctrl, botinput.TypeText), text: original}
	webhookDriver{}.logInput(ctx, 0, input)
	var received string
	command := botsfw.Command{TextAction: func(_ botsfw.WebhookContext, text string) (botmsg.MessageFromBot, error) {
		received = text
		return botmsg.MessageFromBot{}, nil
	}}
	if _, err := command.TextAction(nil, input.Text()); err != nil {
		t.Fatal(err)
	}
	if received != original || input.Text() != original {
		t.Fatal("log redaction changed command input")
	}
	capture.mu.Lock()
	joined := strings.Join(capture.messages, "\n")
	capture.mu.Unlock()
	if strings.Contains(joined, "private-nonce") || !strings.Contains(joined, "watch-[redacted]") {
		t.Fatalf("unsafe log %q", joined)
	}
	previous := RoutingTracer
	t.Cleanup(func() { RoutingTracer = previous })
	var traced string
	RoutingTracer = func(_ context.Context, event RouteTraceEvent) { traced = event.InputText }
	traceRoute(ctx, "start", "command-code", "", original)
	if strings.Contains(traced, "private-nonce") || !strings.Contains(traced, "[redacted]") {
		t.Fatalf("unsafe routing trace %q", traced)
	}
}
