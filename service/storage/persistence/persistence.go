package persistence

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"io"
	"log"
)

type Storage struct {
	S3       *s3.S3
	Uploader *s3manager.Uploader
}

func New(sess *session.Session) *Storage {
	return &Storage{
		S3:       s3.New(sess),
		Uploader: s3manager.NewUploader(sess),
	}
}

func (s *Storage) InitializeS3Bucket(bucketName string) error {
	_, err := s.S3.HeadBucket(
		&s3.HeadBucketInput{
			Bucket: aws.String(bucketName),
		},
	)

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			code := aerr.Code()
			switch code {
			case s3.ErrCodeNoSuchBucket, "NotFound":
				output, err := s.createBucket(bucketName)
				if err != nil {
					return err
				}
				log.Printf("S3 Bucket Initialized at location: %s\n", *output.Location)
			}
		} else {
			return err
		}
	} else {
		log.Println("S3 Bucket already exists. Skipping bucket creation.")
	}

	return nil
}

func (s *Storage) UploadFile(bucketName string, fileName string, body io.Reader) (*s3manager.UploadOutput, error) {
	output, err := s.Uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(fileName),
		Body:   body,
	})
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (s *Storage) createBucket(bucketName string) (*s3.CreateBucketOutput, error) {
	output, err := s.S3.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		return nil, err
	}

	err = s.S3.WaitUntilBucketExists(&s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		return nil, err
	}

	_, err = s.S3.PutBucketVersioning(&s3.PutBucketVersioningInput{
		Bucket: aws.String(bucketName),
		VersioningConfiguration: &s3.VersioningConfiguration{
			Status: aws.String("Enabled"),
		},
	})
	if err != nil {
		return nil, err
	}

	return output, nil
}
