package gae_host

import (
	"testing"
)

const noPanic = "The code did not panic"

func TestGaeBotHost_Logger_nilRequest(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error(noPanic)
		}
	}()

	GaeBotHost{}.Logger(nil)
}

func TestGaeBotHost_GetHttpClient_nilRequest(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error(noPanic)
		}
	}()

	GaeBotHost{}.GetHttpClient(nil)
}

func TestGaeBotHost_GetBotCoreStores_nil(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error(noPanic)
		}
	}()

	GaeBotHost{}.GetBotCoreStores("", nil, nil)
}
