package botswebhook

import (
	"bufio"
	"io"
	"net"
	"net/http"
	"sync/atomic"

	"github.com/felixge/httpsnoop"
)

// MaxWebhookRequestBodyBytes limits webhook payloads before platform adapters
// decode them. Webhook bodies are metadata envelopes rather than media uploads;
// 8 MiB leaves ample room for batched updates while bounding memory use.
const MaxWebhookRequestBodyBytes int64 = 8 << 20

type webhookResponse struct {
	raw       http.ResponseWriter
	writer    http.ResponseWriter
	committed atomic.Bool
}

func newWebhookResponse(w http.ResponseWriter) *webhookResponse {
	response := &webhookResponse{raw: w}
	response.writer = httpsnoop.Wrap(w, httpsnoop.Hooks{
		WriteHeader: func(next httpsnoop.WriteHeaderFunc) httpsnoop.WriteHeaderFunc {
			return func(statusCode int) {
				if statusCode < 100 || statusCode > 199 {
					response.committed.Store(true)
				}
				next(statusCode)
			}
		},
		Write: func(next httpsnoop.WriteFunc) httpsnoop.WriteFunc {
			return func(body []byte) (int, error) {
				response.committed.Store(true)
				return next(body)
			}
		},
		WriteString: func(next httpsnoop.WriteStringFunc) httpsnoop.WriteStringFunc {
			return func(body string) (int, error) {
				response.committed.Store(true)
				return next(body)
			}
		},
		ReadFrom: func(next httpsnoop.ReadFromFunc) httpsnoop.ReadFromFunc {
			return func(source io.Reader) (int64, error) {
				response.committed.Store(true)
				return next(source)
			}
		},
		Flush: func(next httpsnoop.FlushFunc) httpsnoop.FlushFunc {
			return func() {
				response.committed.Store(true)
				next()
			}
		},
		FlushError: func(next httpsnoop.FlushErrorFunc) httpsnoop.FlushErrorFunc {
			return func() error {
				response.committed.Store(true)
				return next()
			}
		},
		Hijack: func(next httpsnoop.HijackFunc) httpsnoop.HijackFunc {
			return func() (net.Conn, *bufio.ReadWriter, error) {
				connection, readWriter, err := next()
				if err == nil {
					response.committed.Store(true)
				}
				return connection, readWriter, err
			}
		},
	})
	return response
}

func (r *webhookResponse) isCommitted() bool {
	return r.committed.Load()
}

// writeError writes a fixed, non-sensitive HTTP error only if no response has
// been committed. It returns false after a responder has already started the
// response, allowing callers to log the failure without a second write.
func (r *webhookResponse) writeError(statusCode int) bool {
	if !r.committed.CompareAndSwap(false, true) {
		return false
	}
	header := r.raw.Header()
	header.Set("Content-Type", "text/plain; charset=utf-8")
	header.Set("X-Content-Type-Options", "nosniff")
	r.raw.WriteHeader(statusCode)
	_, _ = io.WriteString(r.raw, http.StatusText(statusCode)+"\n")
	return true
}
