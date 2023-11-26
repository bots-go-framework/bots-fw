module github.com/bots-go-framework/bots-fw

go 1.21

toolchain go1.21.4

//replace github.com/strongo/app => ../../strongo/app

//replace github.com/strongo/i18n => ../../strongo/i18n

//replace github.com/bots-go-framework/bots-fw-store => ../bots-fw-store

require (
	github.com/bots-go-framework/bots-fw-store v0.4.0
	github.com/bots-go-framework/bots-go-core v0.0.2
	github.com/dal-go/dalgo v0.12.0
	github.com/pquerna/ffjson v0.0.0-20190930134022-aa0246cd15f7
	github.com/stretchr/testify v1.8.4
	github.com/strongo/gamp v0.0.1
	github.com/strongo/i18n v0.0.4
	github.com/strongo/strongoapp v0.9.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/strongo/random v0.0.1 // indirect
	github.com/strongo/validation v0.0.6 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
