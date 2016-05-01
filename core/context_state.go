package bots

import (
	"strings"
)

const (
	AWAITING_REPLY_TO_PATH_SEPARATOR = "/"
	AWAITING_REPLY_TO_PATH2QUERY_SEPARATOR = "?"
	AWAITING_REPLY_TO_PARAMS_SEPARATOR = "&"
)

func AwaitingReplyToPath(awaitingReplyTo string) string {
	s := strings.Split(awaitingReplyTo, AWAITING_REPLY_TO_PATH2QUERY_SEPARATOR)
	return s[0]
}

func AwaitingReplyToQuery(awaitingReplyTo string) string {
	s := strings.Split(awaitingReplyTo, AWAITING_REPLY_TO_PATH2QUERY_SEPARATOR)
	return s[1]
}


