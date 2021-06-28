package sqs

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/bogdanrat/web-server/service/queue"
	"log"
	"strings"
)

type sqsEventEmitter struct {
	svc      *sqs.SQS
	queueUrl *string
	isFifo   bool
	config   Config
}

func NewEventEmitter(sess *session.Session, config Config) (queue.EventEmitter, error) {
	if sess == nil {
		return nil, fmt.Errorf("aws session is nil")
	}

	svc := sqs.New(sess)
	emitter := &sqsEventEmitter{
		svc: svc,
	}

	if err := emitter.setup(config); err != nil {
		return nil, err
	}
	return emitter, nil
}

func (e *sqsEventEmitter) setup(config Config) error {
	output, err := e.svc.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: aws.String(config.QueueName),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			code := aerr.Code()
			switch code {
			case sqs.ErrCodeQueueDoesNotExist, "NotFound":
				if err := e.createQueue(config); err != nil {
					return err
				}
				log.Printf("SQS Queue initialized: %s\n", *e.queueUrl)
				return nil
			}
		}
		return err
	}
	e.queueUrl = output.QueueUrl

	return nil
}

func (e *sqsEventEmitter) createQueue(config Config) error {
	isFifoQueue := "false"
	if strings.Contains(config.QueueName, ".fifo") {
		isFifoQueue = "true"
	}

	output, err := e.svc.CreateQueue(&sqs.CreateQueueInput{
		QueueName: aws.String(config.QueueName),
		Attributes: map[string]*string{
			"FifoQueue":                 aws.String(isFifoQueue),
			"ContentBasedDeduplication": aws.String(config.ContentBasedDeduplication),
			"DelaySeconds":              aws.String(config.DelaySeconds),
			"MessageRetentionPeriod":    aws.String(config.MessageRetentionPeriod),
		},
	})
	if err != nil {
		return err
	}

	e.queueUrl = output.QueueUrl
	return nil
}

func (e *sqsEventEmitter) Emit(event queue.Event) error {
	jsonBody, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("could not marshal json event: %s", err.Error())
	}

	var messageGroupID *string
	switch event.Name() {
	case "userSignUp":
		messageGroupID = aws.String("auth")
	}

	message := &sqs.SendMessageInput{
		QueueUrl:    e.queueUrl,
		MessageBody: aws.String(string(jsonBody)),
		MessageAttributes: map[string]*sqs.MessageAttributeValue{
			"event_name": {
				DataType:    aws.String("string"),
				StringValue: aws.String(event.Name()),
			},
		},
	}
	if e.isFifo {
		message.MessageGroupId = messageGroupID
	}

	_, err = e.svc.SendMessage(message)
	return err
}
