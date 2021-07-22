package sqs

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/bogdanrat/web-server/service/queue"
	"log"
	"strings"
)

type sqsEventListener struct {
	svc      *sqs.SQS
	queueUrl *string
	mapper   queue.EventMapper
	config   Config
}

func NewEventListener(sess *session.Session, config Config) (queue.EventListener, error) {
	if sess == nil {
		return nil, fmt.Errorf("aws session is nil")
	}

	svc := sqs.New(sess)

	mapper, err := queue.NewEventMapper(queue.StaticMapper)
	if err != nil {
		return nil, err
	}

	listener := &sqsEventListener{
		svc:    svc,
		mapper: mapper,
		config: config,
	}

	if err := listener.setup(config); err != nil {
		return nil, err
	}
	return listener, nil
}

func (l *sqsEventListener) setup(config Config) error {
	output, err := l.svc.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: aws.String(config.QueueName),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			code := aerr.Code()
			switch code {
			case sqs.ErrCodeQueueDoesNotExist, "NotFound":
				log.Printf("SQS Queue not found: %s\n", config.QueueName)
				return err
			}
		}
		return err
	}
	l.queueUrl = output.QueueUrl
	return nil
}

func (l *sqsEventListener) Listen(eventNames ...string) (<-chan queue.Event, <-chan error, error) {
	events := make(chan queue.Event)
	errors := make(chan error)

	go func() {
		for {
			l.receiveMessage(events, errors, eventNames...)
		}
	}()

	return events, errors, nil
}

func (l *sqsEventListener) receiveMessage(events chan queue.Event, errors chan error, eventNames ...string) {
	output, err := l.svc.ReceiveMessage(&sqs.ReceiveMessageInput{
		QueueUrl:              l.queueUrl,
		MaxNumberOfMessages:   aws.Int64(l.config.MaxNumberOfMessages),
		VisibilityTimeout:     aws.Int64(l.config.VisibilityTimeout),
		WaitTimeSeconds:       aws.Int64(l.config.WaitTimeSeconds),
		AttributeNames:        aws.StringSlice([]string{sqs.MessageSystemAttributeNameMessageGroupId}),
		MessageAttributeNames: aws.StringSlice([]string{sqs.QueueAttributeNameAll}),
	})
	if err != nil {
		errors <- err
		return
	}

	foundEvent := false
	for _, message := range output.Messages {
		attributeValue, ok := message.MessageAttributes["event_name"]
		if !ok {
			continue
		}
		messageGroupID, ok := message.Attributes[sqs.MessageSystemAttributeNameMessageGroupId]
		if ok && strings.EqualFold(*messageGroupID, MessageGroupIDAuth) {

		}

		eventName := aws.StringValue(attributeValue.StringValue)
		for _, event := range eventNames {
			if strings.EqualFold(eventName, event) {
				foundEvent = true
				break
			}
		}
		if !foundEvent {
			continue
		}

		messageBody := aws.StringValue(message.Body)
		event, err := l.mapper.MapEvent(eventName, []byte(messageBody))
		if err != nil {
			errors <- err
			continue
		}
		events <- event

		_, err = l.svc.DeleteMessage(&sqs.DeleteMessageInput{
			QueueUrl:      l.queueUrl,
			ReceiptHandle: message.ReceiptHandle,
		})
		if err != nil {
			errors <- err
		}
	}
}

func (l *sqsEventListener) EventMapper() queue.EventMapper {
	return l.mapper
}
