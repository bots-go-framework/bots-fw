package botswebhook

import "strings"

// botLogText redacts single-use Watch account-linking bearer payloads only at
// observability boundaries. Routing and command execution receive the original.
func botLogText(text string) string {
	fields := strings.Fields(text)
	if len(fields) >= 2 {
		command := strings.ToLower(fields[0])
		if (command == "/start" || strings.HasPrefix(command, "/start@")) && strings.HasPrefix(fields[1], "watch-") {
			return fields[0] + " watch-[redacted]"
		}
	}
	return text
}
