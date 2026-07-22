package botsfw

import (
	"context"
	"testing"
)

func envProviderFromMap(values map[string]string) EnvProvider {
	return EnvProviderFunc(func(name string) (string, bool) {
		value, ok := values[name]
		return value, ok
	})
}

func TestEnvProviderDefaultsToProcessEnvironment(t *testing.T) {
	const name = "BOTFW_ENV_PROVIDER_DEFAULT_TEST"
	t.Setenv(name, "from-process")

	value, ok := LookupEnv(context.Background(), name)
	if !ok || value != "from-process" {
		t.Fatalf("LookupEnv(%q) = %q, %v; want from-process, true", name, value, ok)
	}
}

func TestWithEnvProviderRejectsNil(t *testing.T) {
	defer func() {
		if recovered := recover(); recovered != "botsfw.WithEnvProvider: nil provider" {
			t.Fatalf("panic = %v, want nil-provider panic", recovered)
		}
	}()
	WithEnvProvider(context.Background(), nil)
}

func TestContextEnvProvidersAreIndependent(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name  string
		value string
	}{
		{name: "first", value: "first-value"},
		{name: "second", value: "second-value"},
	} {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			ctx := WithEnvProvider(context.Background(), envProviderFromMap(map[string]string{
				"SHARED_NAME": test.value,
			}))
			for i := 0; i < 1000; i++ {
				if got := GetEnv(ctx, "SHARED_NAME"); got != test.value {
					t.Fatalf("GetEnv() = %q, want %q", got, test.value)
				}
				if _, ok := LookupEnv(ctx, "MISSING"); ok {
					t.Fatal("LookupEnv(MISSING) unexpectedly reported a value")
				}
			}
		})
	}
}
