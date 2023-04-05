# Strongo Bots Framework
A [Go language](https://golang.org/) framework to develop bots for messengers.

**Reasons to use**:
 
 * Same code can work across different  messenger (_Telegram, Facebook Messenger, Viber, Skype, Line, Kik, WeChat, etc._)
 * You can tune your code to a specific messenger's APIs.
 * i18n & l10n support (_multilingual_)   
 * Can be hosted in cloud or just as a standard Go HTTP server. Supports [AppEngine](https://cloud.google.com/appengine/) standard environment.
 * It's fast   


## Conitious Integration
![Build and Test](https://github.com/strongo/bots-framework/actions/workflows/test.yml/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/strongo/bots-framework?cache=1)](https://goreportcard.com/report/github.com/strongo/bots-framework)
[![GoDoc](https://godoc.org/github.com/strongo/bots-framework?status.svg)](https://godoc.org/github.com/strongo/bots-framework)

## Usage

	func InitBot(httpRouter *httprouter.Router, botHost bots.BotHost, appContext common.DebtsTrackerAppContext) {
	
		driver := bots.NewBotDriver( // Orchestrate requests to appropriate handlers
			bots.AnalyticsSettings{GaTrackingID: common.GA_TRACKING_ID}, // TODO: Refactor to list of analytics providers
			appContext,                                       // Holds User entity kind name, translator, etc.
			botHost,                                          // Defines how to create context.Context, HttpClient, DB, etc...
			"Please report any issues to @DebtsTrackerGroup", // Is it wrong place? Router has similar.
		)
	
		driver.RegisterWebhookHandlers(httpRouter, "/bot",
			telegram.NewTelegramWebhookHandler(
				telegramBotsWithRouter, // Maps of bots by code, language, token, etc...
				newTranslator, // Creates translator that gets a context.Context (for logging purpose)
			),
			viber.NewViberWebhookHandler(
				viber.Bots,
				newTranslator,
			),
			fbm.NewFbmWebhookHandler(
				fbm.Bots,
				newTranslator,
			),
		)
	}

## Sample bots built with Strongo Bots Framework
The best way to learn is to see examples of usage. Here is few:
  * [**Reversi** game](https://github.com/prizarena/reversi) - open source game. (*Telegram: [@ReversiGameBot](https://t.me/ReversiGameBot)*)
  * [**Bidding Tic-Tac-Toe**](https://github.com/prizarena/bidding-tictactoe) - open source game. (*Telegram: [@BiddingTicTacToeBot](https://t.me/BiddingTicTacToeBot)*)
  * [**Rock-Paper-Scissors**](https://github.com/prizarena/rock-paper-scissors) - open source game. (*Telegram: [@playRockPaperScissorsBot](https://t.me/playRockPaperScissorsBot)*)
  * [**Debtus.app**](http://debtus.app/) —  a bot & a reminder service that helps to track your debts & credits.
  Sends automated reminders to you & your debtors (_in messenger, email, SMS_).

We would be happy to place a link to your example / bot that is implemented using this framework.

## Go API libraries used by the framework to talk to messengers
You can use any Bot API library by implementing couple of simple interface but the framework comes with few buildins:
  * [strongo/bots-api-telegram](https://github.com/strongo/bots-api-telegram) - Go library for [**Telegram** Bot API](https://core.telegram.org/bots/api)
  * [strongo/bots-api-fbm](https://github.com/strongo/bots-api-fbm) - Go library for [**Facebook Messenger** Bot API](https://developers.facebook.com/docs/messenger-platform)
  * [strongo/bots-api-viber](https://github.com/strongo/bots-api-viber) - Go library for [**Viber** Bot API](https://developers.viber.com/)
  
## Other Go libraries used by the bot framework
  * [strongo/dalgo](https://github.com/dal-go/dalgo) - Database abstraction layer (DAL) in Go language
  * [strongo/gamp](https://github.com/strongo/gamp) - Golang buffered client for Google Analytics (GA) Measurement Protocol
  * [julienschmidt/httprouter](https://github.com/julienschmidt/httprouter) - A high performance HTTP request router that scales well
  * [pkg/errors](https://github.com/pkg/errors) - Simple error handling primitives

## [Can I use &mdash; features cross-table for bot messenger APIs](can-i-use-bots-api.md)
We are building a [cross-table of features](can-i-use-bots-api.md) supported by different bot APIs.
  
## Contributors
  * [Alexander Trakhimenok](https://ie.linkedin.com/in/alexandertrakhimenok)

## Press
There are no articles about the Strongo Bots Framework just yet. Send us a link if you find such.
  
## [License](https://github.com/strongo/bots-framework/blob/master/LICENSE)
Licensed under Apache 2.0 license
