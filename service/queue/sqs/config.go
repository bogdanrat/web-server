package sqs

const (
	MessageGroupIDAuth = "auth"
)

type Config struct {
	QueueName                 string
	ContentBasedDeduplication string
	DelaySeconds              string
	MessageRetentionPeriod    string
	MaxNumberOfMessages       int64
	VisibilityTimeout         int64
	WaitTimeSeconds           int64
}
