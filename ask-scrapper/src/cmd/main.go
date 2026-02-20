package main

import (
	"fmt"
	"log"

	"ask-scrapper/src/internal/config"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	fmt.Printf("Подключение к RabbitMQ: %s\n", cfg.RabbitMQ.URL)

	conn, err := amqp.Dial(cfg.RabbitMQ.URL)
	if err != nil {
		log.Fatalf("Ошибка подключения к RabbitMQ: %v", err)
	}
	defer conn.Close()

	fmt.Println("Успешно подключено к RabbitMQ!")

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Ошибка открытия канала: %v", err)
	}
	defer ch.Close()

	fmt.Println("Канал открыт!")
}
