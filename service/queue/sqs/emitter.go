package sqs

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/bogdanrat/web-server/contracts/models"
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
	if strings.Contains(config.QueueName, ".fifo") {
		e.isFifo = true
	}

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
			}
		}
		return err
	}
	e.queueUrl = output.QueueUrl

	return nil
}

func (e *sqsEventEmitter) createQueue(config Config) error {
	createQueueInput := &sqs.CreateQueueInput{
		QueueName: aws.String(config.QueueName),
		Attributes: map[string]*string{
			"DelaySeconds":           aws.String(config.DelaySeconds),
			"MessageRetentionPeriod": aws.String(config.MessageRetentionPeriod),
		},
	}
	if e.isFifo {
		createQueueInput.Attributes["FifoQueue"] = aws.String("true")
		createQueueInput.Attributes["ContentBasedDeduplication"] = aws.String(config.ContentBasedDeduplication)
	}

	output, err := e.svc.CreateQueue(createQueueInput)
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
	case models.UserSignUpEventName:
		messageGroupID = aws.String(MessageGroupIDAuth)
	}

	message := &sqs.SendMessageInput{
		QueueUrl:    e.queueUrl,
		MessageBody: aws.String(string(jsonBody)),
		MessageAttributes: map[string]*sqs.MessageAttributeValue{
			queue.EventNameHeader: {
				DataType:    aws.String("String"),
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
