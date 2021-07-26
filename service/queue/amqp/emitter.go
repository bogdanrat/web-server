package amqp

import (
	"encoding/json"
	"fmt"
	"github.com/bogdanrat/web-server/service/queue"
	"github.com/streadway/amqp"
)

/*
	* Publish/Subscribe Pattern:
		- publishers and subscribers are very LOOSELY COUPLED: they do not even know about one another.
		- FLEXIBLE: add new subscribers/publishers without modifying the publishers/subscribers
		- RESILIENCY: the message broker stores all messages in a queue, in which they are kept until they have been processed by a subscriber;
		if a subscriber becomes unavailable, the messages that should have been routed to that subscriber will become queued until the subscriber is available again.
		- RELIABLE DELIVERY: require each subscriber to acknowledge a received message; only when the message has been acknowledged, the broker will remove the message from the queue.
		- SCALE OUT EASILY: add more subscribers for a publishers and have the message broker load-balance the messages sent to the subscribers.

	* AMQP (Advanced Message Queueing Protocol)
		- Exchange: each publisher publishes its messages into an exchange.
		- Queue: each subscriber consumes a queue.
		Where messages fo after they have been published to an exchange depends on the EXCHANGE TYPE and the routing rules called BINDINGS:
			- DIRECT EXCHANGES: messages are published with a given topic (routing key in AMQP) that is a simple string value.
			- FANOUT EXCHANGES: messages are routed to all queues that are connected to a fanout exchange via a binding.
			- TOPIC EXCHANGES: similar to direct exchanges, but queues are bound to the exchange using patterns that the message's routing key must match.
			(<entity-name>.<state-change>.<location>, e.g., event.created.europe; event.created.*; event.*.europe)
*/

type amqpEventEmitter struct {
	connection *amqp.Connection
	exchange   string
	events     chan *emittedEvent
}

type emittedEvent struct {
	event     queue.Event
	errorChan chan error
}

func NewEventEmitter(conn *amqp.Connection, exchange string) (queue.EventEmitter, error) {
	emitter := &amqpEventEmitter{
		connection: conn,
		exchange:   exchange,
	}

	err := emitter.setup()
	if err != nil {
		return nil, err
	}
	return emitter, nil
}

func (e *amqpEventEmitter) setup() error {
	// Channels are used to multiplex several virtual connections over one actual TCP connection.
	// Channel() opens a unique, concurrent server channel to process the bulk of AMQP messages
	channel, err := e.connection.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()

	// A message publishers declares the Exchange into which it intends to publish messages
	err = channel.ExchangeDeclare(
		e.exchange,
		"topic",
		true,  // durable: cause the exchange to remain declared when the broker restarts
		false, // autoDelete: cause the exchange to be deleted as soon as the channel that declared it is closed
		false, // internal: prevent publishers from publishing messages into this queue
		false, // noWait: instruct the ExchangeDeclare method not to wait for a successful response from the broker
		nil,
	)
	return err
}

func (e *amqpEventEmitter) Emit(event queue.Event) error {
	channel, err := e.connection.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()

	jsonBody, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("could not marshal json event: %s", err.Error())
	}

	message := amqp.Publishing{
		Headers:     amqp.Table{queue.EventNameHeader: event.Name()},
		ContentType: "application/json",
		Body:        jsonBody,
	}

	err = channel.Publish(
		e.exchange,
		event.Name(),
		false, // mandatory: instruct the broker to make sure that the message is actually routed into at least one queue
		false, // immediate: instruct the broker to make sure that the message is actually delivered to at least on subscriber
		message,
	)

	return err
}
