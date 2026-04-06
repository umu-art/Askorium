module ask-crawler

go 1.25.0

require (
	github.com/omo-ri/askorium/go-crawler-api v0.0.0-00010101000000-000000000000
	github.com/omo-ri/askorium/go-parser-api v0.0.0-00010101000000-000000000000
	github.com/omo-ri/askorium/go-renderer-api v0.0.0-00010101000000-000000000000
	github.com/rabbitmq/amqp091-go v1.10.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/redis/go-redis/v9 v9.18.0 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	gopkg.in/validator.v2 v2.0.1 // indirect
)

replace (
	github.com/omo-ri/askorium/go-crawler-api => ./../api/build/go-crawler-api
	github.com/omo-ri/askorium/go-parser-api => ./../api/build/go-parser-api
	github.com/omo-ri/askorium/go-renderer-api => ./../api/build/go-renderer-api
)
