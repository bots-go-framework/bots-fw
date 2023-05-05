package botsfw

import "context"

func InitializeBotsFw(logger Logger) {
	if logger == nil {
		panic("logger is nil")
	}
	log = logger
	log.Infof(context.Background(), "botsfw initialized")
}
