# Strongo Bots Framework
This is a [Go language](https://golang.org/) framework for building multilingual messenger bots (_Telegram, Facebook Messenger, Skype, Line, Kik, WeChat_) hosted on [AppEngine](https://cloud.google.com/appengine/), [Amazon](https://aws.amazon.com/), [Azure](https://azure.microsoft.com/), [Heroku](https://www.heroku.com/), [Docker](https://www.docker.com/) or just as a standard Go HTTP server.

## Sample bots built with Strongo Bots Framework
The best way to learn is to see examples of usage. Here is few:
  * [strongo/bots-example-calculator](http://github.com/strongo/bots-example-calculator) — a simple bot that calculates math expressions
  * [strongo/bots-example-rock-paper-scissors](http://github.com/strongo/bots-example-rock-paper-scissors) — a bot to play [Rock-Paper-Scissors](https://en.wikipedia.org/wiki/Rock-paper-scissors) with your friends or agains AI
  * [**DebtsTracker.io**](http://debtstracker.io/) —  a bot & reminder service that helps to track your debts & credits. Sends automated email & SMS notifications to your debtors.

## Go API libraries used by the framework to talk to messengers
You can use any Bot API library by implementing couple of simple interface but the framework comes with few buildins:
  * [strongo/bots-api-telegram](strongo/bots-api-telegram) - Go library for [**Telegram** Bot API](https://core.telegram.org/bots/api)
  * [strongo/bots-api-fbm](strongo/bots-api-fbm) - Go library for [**Facebook Messenger** Bot API](https://developers.facebook.com/docs/messenger-platform)
  
