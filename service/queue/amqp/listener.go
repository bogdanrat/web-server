package amqp

import (
	"encoding/json"
	"fmt"
	"github.com/bogdanrat/web-server/contracts/models"
	"github.com/bogdanrat/web-server/service/queue"
	"github.com/streadway/amqp"
)

type amqpEventListener struct {
	connection *amqp.Connection
	exchange   string
	queue      string
}

func NewListener(conn *amqp.Connection, exchange string, queue string) (queue.EventListener, error) {
	listener := &amqpEventListener{
		connection: conn,
		exchange:   exchange,
		queue:      queue,
	}

	if err := listener.setup(); err != nil {
		return nil, err
	}

	return listener, nil
}

func (l *amqpEventListener) Listen(eventNames ...string) (<-chan queue.Event, <-chan error, error) {
	channel, err := l.connection.Channel()
	if err != nil {
		return nil, nil, err
	}
	defer channel.Close()

	for _, eventName := range eventNames {
		// QueueBind binds an exchange to a queue so that publishings to the exchange will be routed to the queue
		// when the publishing routing key matches the binding routing key
		if err := channel.QueueBind(l.queue, eventName, l.exchange, false, nil); err != nil {
			return nil, nil, err
		}
	}

	// Consume() immediately starts delivering queued messages.
	messages, err := channel.Consume(
		l.queue,
		"",    // consumer: when empty, a unique identifier will be automatically generated
		false, // autoAck: when true, received messages will be acknowledged automatically; when false, use Ack() method
		false, // exclusive: when true, this consumer will be the only one allowed to consume this queue
		false, // noLocal: this consumer should not be delivered messages that were published on the same channel
		false, // noWait: instructs the library not to wait for confirmation from the broker
		nil,
	)
	if err != nil {
		return nil, nil, err
	}

	events := make(chan queue.Event)
	errors := make(chan error)

	go func() {
		for message := range messages {
			// use the x-event-name header to map message back to their respective struct types
			rawEventName, ok := message.Headers["x-event-name"]
			if !ok {
				errors <- fmt.Errorf("message did not contain x-event-name header")
				// Nack() negatively acknowledge the delivery of message(s)
				// This method must not be used to select or requeue messages the client wishes not to handle,
				// rather it is to inform the server that the client is incapable of handling this message at this time.

				// When requeue is true, request the server to deliver this message to a different consumer.
				// If it is not possible or requeue is false, the message will be dropped or delivered to a server configured dead-letter queue.
				message.Nack(false, false)
				continue
			}

			// todo event mapper, see cloud native w. go book

			var event queue.Event
			switch rawEventName {
			case "userSignUp":
				event = new(models.UserSignUp)
			default:
				errors <- fmt.Errorf("event type %s is unknown", rawEventName)
				continue
			}

			err := json.Unmarshal(message.Body, event)
			if err != nil {
				errors <- err
				continue
			}

			events <- event
			err = message.Ack(false)
			if err != nil {
				errors <- fmt.Errorf("could not acknowledge message: %s", err)
			}
		}
	}()

	return events, errors, nil
}

func (l *amqpEventListener) setup() error {
	channel, err := l.connection.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()

	// QueueDeclare declares a queue to hold messages and deliver to consumers
	_, err = channel.QueueDeclare(l.queue,
		true,
		false,
		false,
		false,
		nil,
	)
	return err
}
