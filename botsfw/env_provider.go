package botsfw

import (
	"context"
	"os"
)

// EnvProvider resolves immutable runtime configuration by name.
// Implementations must be safe for concurrent use when the same provider is
// shared by concurrent request contexts.
type EnvProvider interface {
	LookupEnv(name string) (value string, ok bool)
}

// EnvProviderFunc adapts a function to EnvProvider.
type EnvProviderFunc func(name string) (value string, ok bool)

// LookupEnv implements EnvProvider.
func (f EnvProviderFunc) LookupEnv(name string) (value string, ok bool) {
	return f(name)
}

type osEnvProvider struct{}

func (osEnvProvider) LookupEnv(name string) (value string, ok bool) {
	return os.LookupEnv(name)
}

type envProviderContextKey struct{}

var defaultEnvProvider EnvProvider = osEnvProvider{}

// WithEnvProvider returns a child context that resolves environment-backed
// runtime configuration through provider. It does not mutate process state.
func WithEnvProvider(ctx context.Context, provider EnvProvider) context.Context {
	if provider == nil {
		panic("botsfw.WithEnvProvider: nil provider")
	}
	return context.WithValue(ctx, envProviderContextKey{}, provider)
}

// EnvProviderFromContext returns the context provider, or the process
// environment provider when no override was attached.
func EnvProviderFromContext(ctx context.Context) EnvProvider {
	if ctx != nil {
		if provider, ok := ctx.Value(envProviderContextKey{}).(EnvProvider); ok && provider != nil {
			return provider
		}
	}
	return defaultEnvProvider
}

// LookupEnv resolves name through the provider attached to ctx.
func LookupEnv(ctx context.Context, name string) (value string, ok bool) {
	return EnvProviderFromContext(ctx).LookupEnv(name)
}

// GetEnv resolves name through the provider attached to ctx and returns an
// empty string when the name is not present.
func GetEnv(ctx context.Context, name string) string {
	value, _ := LookupEnv(ctx, name)
	return value
}
