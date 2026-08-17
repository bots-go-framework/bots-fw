module github.com/bots-go-framework/bots-fw

go 1.25

//replace github.com/strongo/app => ../../strongo/app
//replace github.com/strongo/i18n => ../../strongo/i18n
//replace github.com/bots-go-framework/bots-go-core => ../bots-go-core
require (
	github.com/bots-go-framework/bots-fw-store v0.14.1
	github.com/bots-go-framework/bots-go-core v0.3.0
	github.com/felixge/httpsnoop v1.1.0
	github.com/stretchr/testify v1.12.0
	github.com/strongo/analytics v0.2.5
	github.com/strongo/i18n v0.8.16
	github.com/strongo/logus v0.4.1
	github.com/strongo/validation v0.0.10
	go.uber.org/mock v0.6.0
)

require github.com/bots-go-framework/bots-api-telegram v0.15.16

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rogpeppe/go-internal v1.9.0 // indirect
	github.com/strongo/slice v0.3.5 // indirect
	github.com/technoweenie/multipartstreamer v1.0.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
