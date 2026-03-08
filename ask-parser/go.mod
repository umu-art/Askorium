module ask-parser

go 1.25.7

require (
	github.com/PuerkitoBio/goquery v1.11.0
	github.com/omo-ri/askorium/go-renderer-api v0.0.0
	github.com/omo-ri/askorium/go-scrapper-api v0.0.0
	github.com/rabbitmq/amqp091-go v1.10.0
)

require (
	github.com/andybalholm/cascadia v1.3.3 // indirect
	golang.org/x/net v0.47.0 // indirect
	gopkg.in/validator.v2 v2.0.1 // indirect
)

replace (
	github.com/omo-ri/askorium/go-renderer-api => ./../api/build/go-renderer-api
	github.com/omo-ri/askorium/go-scrapper-api => ./../api/build/go-scrapper-api
)
