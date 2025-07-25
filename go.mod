module github.com/bots-go-framework/bots-fw

go 1.24.3

//replace github.com/strongo/app => ../../strongo/app
//replace github.com/strongo/i18n => ../../strongo/i18n
//replace github.com/bots-go-framework/bots-go-core => ../bots-go-core
//replace github.com/bots-go-framework/bots-fw-store => ../bots-fw-store

require (
	github.com/bots-go-framework/bots-fw-store v0.10.0
	github.com/bots-go-framework/bots-go-core v0.2.3
	github.com/dal-go/dalgo v0.24.0
	github.com/stretchr/testify v1.10.0
	github.com/strongo/analytics v0.2.2
	github.com/strongo/i18n v0.8.5
	github.com/strongo/logus v0.2.1
	go.uber.org/mock v0.5.2
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/strongo/random v0.0.1 // indirect
	github.com/strongo/slice v0.3.1 // indirect
	github.com/strongo/validation v0.0.7 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
