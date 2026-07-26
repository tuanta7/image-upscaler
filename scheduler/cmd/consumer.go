package main

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/tuanta7/image-upscaler/scheduler/internal/upscale"
)

func initConsumer(conn *amqp.Connection) (<-chan amqp.Delivery, error) {
	channel, err := conn.Channel()
	if err != nil {
		log.Printf("open channel: %v", err)
		return nil, err
	}

	results, err := channel.Consume(
		upscale.ResultsQueue, "",
		true, false, false, false,
		nil,
	)
	if err != nil {
		log.Printf("consume results queue: %v", err)
		return nil, err
	}

	return results, nil
}
