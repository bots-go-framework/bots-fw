package botswebhook

import (
	"errors"
	"fmt"
	"html"
	"net/http"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/bots-go-framework/bots-api-telegram/tgbotapi"
	"github.com/bots-go-framework/bots-fw/botmsg"
	"github.com/bots-go-framework/bots-fw/botsfw"
	"github.com/bots-go-framework/bots-fw/botsfwconst"
)

const maxUserTechnicalDetailsBytes = 2800

var sensitiveErrorPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\b(authorization|proxy-authorization)\s*[:=]\s*(bearer\s+)?[^\s,;]+`),
	regexp.MustCompile(`(?i)\b(bot[_ -]?token|token|password|passwd|secret|credential|api[_ -]?key)\s*[:=]\s*("[^"]*"|'[^']*'|[^\s,;]+)`),
	regexp.MustCompile(`(?i)\b(raw[_ -]?)?(http[_ -]?)?request[_ -]?body\s*[:=]\s*.*`),
	regexp.MustCompile(`(?i)\b(wallet|game)[_ -]?private[_ -]?state\s*[:=]\s*.*`),
	regexp.MustCompile(`(?i)\b(private[_ -]?state)\s*[:=]\s*.*`),
	regexp.MustCompile(`(?i)https?://[^/\s:@]+:[^/\s@]+@`),
}

type userErrorCopy struct {
	technicalDetails      string
	providerRejected      string
	reason                string
	method                string
	code                  string
	description           string
	correlationID         string
	errorChain            string
	genericFailureSummary string
}

func userErrorCopyFor(localeCode5 string) userErrorCopy {
	if strings.HasPrefix(strings.ToLower(localeCode5), "ru") {
		return userErrorCopy{
			technicalDetails:      "🔎 Технические подробности",
			providerRejected:      "⚠️ Telegram отклонил это сообщение, поэтому оно не было доставлено.",
			reason:                "Причина",
			method:                "Метод Telegram",
			code:                  "Код ошибки Telegram",
			description:           "Описание Telegram",
			correlationID:         "ID запроса",
			errorChain:            "Цепочка ошибок",
			genericFailureSummary: "💢 Ошибка сервера — не удалось обработать сообщение.",
		}
	}
	return userErrorCopy{
		technicalDetails:      "🔎 Technical details",
		providerRejected:      "⚠️ Telegram rejected this message, so it wasn’t delivered.",
		reason:                "Reason",
		method:                "Telegram method",
		code:                  "Telegram error code",
		description:           "Telegram description",
		correlationID:         "Request ID",
		errorChain:            "Error chain",
		genericFailureSummary: "💢 Server error — failed to process message.",
	}
}

func expandableUserErrorMessage(whc botsfw.WebhookContext, err error, footer string) (botmsg.MessageFromBot, bool) {
	settings := whc.GetBotSettings()
	if settings == nil ||
		settings.UserErrorDetails.Disclosure != botsfw.UserErrorDetailsDisclosureExpandable ||
		settings.Platform != botsfwconst.PlatformTelegram {
		return botmsg.MessageFromBot{}, false
	}

	copy := userErrorCopyFor(whc.Locale().Code5)
	providerDetails, isProviderError := tgbotapi.TelegramProviderErrorDetailsFrom(err)
	isProviderRejection := isProviderError &&
		providerDetails.ErrorCode >= http.StatusBadRequest &&
		providerDetails.ErrorCode < http.StatusInternalServerError

	var visibleLines []string
	if isProviderRejection {
		visibleLines = append(visibleLines, copy.providerRejected)
		reason := conciseProviderReason(providerDetails.Description)
		if reason != "" {
			visibleLines = append(visibleLines, copy.reason+": "+reason)
		}
	} else {
		visibleLines = append(
			visibleLines,
			whc.Translate(botsfw.MessageTextOopsSomethingWentWrong),
			copy.genericFailureSummary,
		)
	}
	if footer != "" {
		visibleLines = append(visibleLines, footer)
	}

	details := technicalErrorDetails(whc, err, providerDetails, isProviderError, copy)
	message := whc.NewMessage(
		html.EscapeString(strings.Join(visibleLines, "\n\n")) +
			"\n\n<blockquote expandable><b>" + html.EscapeString(copy.technicalDetails) +
			"</b>\n<code>" + html.EscapeString(details) + "</code></blockquote>",
	)
	message.Format = botmsg.FormatHTML
	return message, true
}

func conciseProviderReason(description string) string {
	description = strings.TrimSpace(strings.Split(description, "\n")[0])
	if prefix, reason, found := strings.Cut(description, ":"); found &&
		strings.EqualFold(strings.TrimSpace(prefix), "bad request") {
		description = strings.TrimSpace(reason)
	}
	const maxReasonBytes = 160
	return truncateUTF8(description, maxReasonBytes)
}

func technicalErrorDetails(
	whc botsfw.WebhookContext,
	err error,
	providerDetails tgbotapi.TelegramProviderErrorDetails,
	hasProviderDetails bool,
	copy userErrorCopy,
) string {
	var lines []string
	if hasProviderDetails {
		// Telegram's structured provider description is deliberately shown
		// verbatim under this explicit opt-in policy. The API adapter excludes
		// the token, raw request/response bodies, and request parameters from
		// these details.
		lines = append(
			lines,
			fmt.Sprintf("%s: %s", copy.method, providerDetails.Method),
			fmt.Sprintf("%s: %d", copy.code, providerDetails.ErrorCode),
			fmt.Sprintf("%s: %s", copy.description, providerDetails.Description),
		)
	}
	if requestID := requestCorrelationID(whc.Request()); requestID != "" {
		lines = append(lines, copy.correlationID+": "+redactUserErrorText(requestID, whc.GetBotSettings()))
	}
	lines = append(lines, copy.errorChain+":")
	for i, message := range errorChain(err) {
		lines = append(
			lines,
			fmt.Sprintf("%d. %s", i+1, redactUserErrorText(message, whc.GetBotSettings())),
		)
	}

	return truncateUTF8(strings.Join(lines, "\n"), maxUserTechnicalDetailsBytes)
}

func redactUserErrorText(value string, settings *botsfw.BotSettings) string {
	redactions := []string{
		settings.Token,
		settings.PaymentToken,
		settings.PaymentTestToken,
		settings.VerifyToken,
		settings.GAToken,
		settings.WebhookSecretToken,
	}
	for _, secret := range redactions {
		if secret = strings.TrimSpace(secret); secret != "" {
			value = strings.ReplaceAll(value, secret, "[REDACTED]")
		}
	}
	for _, pattern := range sensitiveErrorPatterns {
		value = pattern.ReplaceAllStringFunc(value, redactSensitiveMatch)
	}
	return value
}

func redactSensitiveMatch(match string) string {
	if strings.HasPrefix(strings.ToLower(match), "http://") ||
		strings.HasPrefix(strings.ToLower(match), "https://") {
		schemeEnd := strings.Index(match, "://") + len("://")
		return match[:schemeEnd] + "[REDACTED]@"
	}
	for _, separator := range []string{":", "="} {
		if i := strings.Index(match, separator); i >= 0 {
			return match[:i+1] + " [REDACTED]"
		}
	}
	return "[REDACTED]"
}

func errorChain(err error) []string {
	const maxDepth = 16
	var messages []string
	var visit func(error, int)
	visit = func(current error, depth int) {
		if current == nil || depth >= maxDepth {
			return
		}
		messages = append(messages, current.Error())
		switch unwrapped := current.(type) {
		case interface{ Unwrap() []error }:
			for _, cause := range unwrapped.Unwrap() {
				visit(cause, depth+1)
			}
		default:
			visit(errors.Unwrap(current), depth+1)
		}
	}
	visit(err, 0)
	return messages
}

func requestCorrelationID(request *http.Request) string {
	if request == nil {
		return ""
	}
	for _, header := range []string{
		"X-Request-ID",
		"X-Correlation-ID",
		"Traceparent",
		"X-Cloud-Trace-Context",
	} {
		if value := strings.TrimSpace(request.Header.Get(header)); value != "" {
			return value
		}
	}
	return ""
}

func truncateUTF8(value string, maxBytes int) string {
	if len(value) <= maxBytes {
		return value
	}
	value = value[:maxBytes]
	for !utf8.ValidString(value) {
		value = value[:len(value)-1]
	}
	return strings.TrimRight(value, " \n\t") + "…"
}
