package bots

import (
	"testing"
	"github.com/strongo/app"
)

func TestNewBotSettings(t *testing.T) {
	const (
		code        = "TestBot1"
		token       = "TestToken2"
		localeCode5 = "Kode5"
	)
	bs := NewBotSettings(Development, code, token, strongo.Locale{Code5: localeCode5})
	if bs.Code != code {
		t.Errorf("Unexpected code: %v", bs.Code)
	}
	if bs.Token != token {
		t.Errorf("Unexpected token: %v", bs.Token)
	}
	if bs.Locale.Code5 != localeCode5 {
		t.Errorf("Unexpected strongo.Locale: %v", bs.Locale.Code5)
	}
}
