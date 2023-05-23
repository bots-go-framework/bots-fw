module github.com/bots-go-framework/bots-fw

go 1.20

//replace github.com/strongo/app => ../../strongo/app
//
//replace github.com/strongo/i18n => ../../strongo/i18n

//replace github.com/bots-go-framework/bots-fw-store => ../bots-fw-store

require (
	github.com/bots-go-framework/bots-fw-store v0.0.7
	github.com/bots-go-framework/bots-go-core v0.0.1
	github.com/pquerna/ffjson v0.0.0-20190930134022-aa0246cd15f7
	github.com/stretchr/testify v1.8.3
	github.com/strongo/app v0.5.4
	github.com/strongo/gamp v0.0.1
	github.com/strongo/i18n v0.0.4
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/strongo/validation v0.0.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
