module github.com/bots-go-framework/bots-fw

go 1.20

//replace github.com/strongo/app => ../../strongo/app
//
//replace github.com/strongo/i18n => ../../strongo/i18n

//replace github.com/bots-go-framework/bots-fw-models => ../bots-fw-models

require (
	github.com/bots-go-framework/bots-fw-models v0.0.2
	github.com/bots-go-framework/bots-go-core v0.0.1
	github.com/dal-go/dalgo v0.2.26
	github.com/pquerna/ffjson v0.0.0-20190930134022-aa0246cd15f7
	github.com/stretchr/testify v1.8.2
	github.com/strongo/app v0.4.0
	github.com/strongo/gamp v0.0.1
	github.com/strongo/i18n v0.0.4
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/strongo/random v0.0.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
