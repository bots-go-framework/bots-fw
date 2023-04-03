package botsfw

import (
	"testing"
)

var testCmd = Command{
	Title: "Title1",
	Titles: map[string]string{
		ShortTitle: "ttl1",
	},
}

var testWhc = &TestWebhookContext{}

func TestCommand_DefaultTitle(t *testing.T) {
	if testCmd.DefaultTitle(testWhc) != "Title1" {
		t.Error("Wrong title")
	}
}

func TestCommand_TitleByKey(t *testing.T) {
	if testCmd.TitleByKey(ShortTitle, testWhc) != "ttl1" {
		t.Error("Wrong title")
	}
}
