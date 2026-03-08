module ask-crawler

go 1.25.0

require (
	github.com/omo-ri/askorium/go-crawler-api v0.0.0-00010101000000-000000000000
	github.com/omo-ri/askorium/go-renderer-api v0.0.0-00010101000000-000000000000
	github.com/omo-ri/askorium/go-parser-api v0.0.0-00010101000000-000000000000
	github.com/rabbitmq/amqp091-go v1.10.0
)

require gopkg.in/validator.v2 v2.0.1 // indirect

replace (
	github.com/omo-ri/askorium/go-crawler-api => ./../api/build/go-crawler-api
	github.com/omo-ri/askorium/go-renderer-api => ./../api/build/go-renderer-api
	github.com/omo-ri/askorium/go-parser-api => ./../api/build/go-parser-api
)
