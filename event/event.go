package event

import (
	"context"
	"encoding/json"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func Setup(connectionString string, queueName string) (chan FindingUpdate, error) {
	connection, err := amqp.Dial(connectionString)
	if err != nil {
		return nil, err
	}
	channel, err := connection.Channel()
	if err != nil {
		return nil, err
	}
	queue, err := channel.QueueDeclare(
		queueName, // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return nil, err
	}
	findingUpdatesChan := make(chan FindingUpdate)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		for findingUpdate := range findingUpdatesChan {
			encoded, err := json.Marshal(findingUpdate)
			if err != nil {
				log.Printf("Failed to marshal finding update: %v... Continued", err)
				continue
			}
			err = channel.PublishWithContext(ctx,
				"",         // exchange
				queue.Name, // routing key
				false,      // mandatory
				false,      // immediate
				amqp.Publishing{
					ContentType: "application/json",
					Body:        encoded,
				},
			)
			if err != nil {
				log.Printf("Failed to publish finding update: %v... Continued", err)
				continue
			}
		}
	}()
	return findingUpdatesChan, nil
}
