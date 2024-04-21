package event

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Kaese72/riskie-lib/logging"

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
		logging.Info(context.Background(), "Started event sender")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		for findingUpdate := range findingUpdatesChan {
			logging.Info(context.Background(), "Received finding update", map[string]interface{}{"findingId": findingUpdate.ID})
			encoded, err := json.Marshal(findingUpdate)
			if err != nil {
				logging.Info(context.Background(), "Failed to marshal finding update... Continued", map[string]interface{}{"findingId": findingUpdate.ID, "error": err.Error()})
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
				logging.Info(context.Background(), "Failed to marshal finding update... Continued", map[string]interface{}{"findingId": findingUpdate.ID, "error": err.Error()})
				continue
			}
		}
	}()
	return findingUpdatesChan, nil
}
