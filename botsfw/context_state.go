package botsfw

import (
	"strings"
)

const (
	// AwaitingReplyToPathSeparator separates parts of the command state
	AwaitingReplyToPathSeparator = "/"

	// AwaitingReplyToPath2QuerySeparator separates path and query parts of state
	AwaitingReplyToPath2QuerySeparator = "?"

	// AwaitingReplyToParamsSeparator separates params of command state
	AwaitingReplyToParamsSeparator = "&"
)

// AwaitingReplyToPath returns just path part of command state
func AwaitingReplyToPath(awaitingReplyTo string) string {
	s := strings.Split(awaitingReplyTo, AwaitingReplyToPath2QuerySeparator)
	return s[0]
}

// AwaitingReplyToQuery returns just query part of command state
func AwaitingReplyToQuery(awaitingReplyTo string) string {
	s := strings.Split(awaitingReplyTo, AwaitingReplyToPath2QuerySeparator)
	if len(s) > 1 {
		return s[1]
	}
	return ""
}
