package main

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/tuanta7/image-upscaler/scheduler/internal/upscale"
)

func declareQueue(ch *amqp.Channel, queueName string) error {
	_, err := ch.QueueDeclare(queueName,
		true,  // Metadata of a durable queue is stored on disk
		false, // Persist even when last consumer unsubscribes
		false, // Persist even when the only one connection close
		false, // waits for the broker's reply
		nil,
	)
	return err
}

func consumeResults(conn *amqp.Connection, handler func([]byte)) error {
	channel, err := conn.Channel()
	if err != nil {
		log.Printf("open channel: %v", err)
		return err
	}

	results, err := channel.Consume(
		upscale.ResultsQueue, "",
		true, false, false, false,
		nil,
	)
	if err != nil {
		log.Printf("consume results queue: %v", err)
		return err
	}

	for msg := range results {
		handler(msg.Body)
	}

	return nil
}
