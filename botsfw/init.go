package botsfw

func InitBotsFrameworkLogger(logger Logger) {
	if logger == nil {
		panic("logger is nil")
	}
	log = logger
}
