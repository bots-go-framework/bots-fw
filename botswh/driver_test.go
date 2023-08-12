package botswh

import "testing"

func TestNewBotDriver(t *testing.T) {
	t.Run("panics_with_bil_app_context", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("The code did not panic")
			} else if err, ok := r.(string); !ok {
				t.Errorf("Expected string, got: %T=%v", r, r)
			} else if err != "appContext == nil" {
				t.Errorf("Unexpected error, got: %v", err)
			}
		}()
		NewBotDriver(AnalyticsSettings{}, nil, nil, "")
	})
}
