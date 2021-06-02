package persistence

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	pb "github.com/bogdanrat/web-server/contracts/proto/storage_service"
	"github.com/bogdanrat/web-server/service/storage/config"
	"github.com/bogdanrat/web-server/service/storage/lib"
	"io"
	"log"
	"sync"
	"time"
)

type Storage struct {
	S3           *s3.S3
	Config       config.S3Config
	UploaderPool sync.Pool
}

func New(sess *session.Session, s3Config config.S3Config) *Storage {
	return &Storage{
		S3:     s3.New(sess),
		Config: s3Config,
		UploaderPool: sync.Pool{
			New: func() interface{} {
				uploader := s3manager.NewUploader(sess)
				// The number of goroutines to spin up in parallel per call to Upload when sending parts
				uploader.Concurrency = s3Config.UploaderConcurrency
				return uploader
			},
		},
	}
}

func (s *Storage) InitializeS3Bucket() error {
	_, err := s.S3.HeadBucket(
		&s3.HeadBucketInput{
			Bucket: aws.String(s.Config.Bucket),
		},
	)

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			code := aerr.Code()
			switch code {
			case s3.ErrCodeNoSuchBucket, "NotFound":
				output, err := s.createBucket()
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

func (s *Storage) UploadFile(key string, body io.Reader) (*s3manager.UploadOutput, error) {
	// Get() first checks if there are any available instances within the pool to return. If not, calls New() to create a new one.
	uploader := s.UploaderPool.Get().(*s3manager.Uploader)
	// Put(): place the instance back in the pool for use by other processes.
	defer s.UploaderPool.Put(uploader)

	fileKey := ""
	if lib.IsImage(key) {
		fileKey = fmt.Sprintf("%s/%s", s.Config.ImagesPrefix, key)
	}

	output, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s.Config.Bucket),
		Key:    aws.String(fileKey),
		Body:   body,
	})
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (s *Storage) createBucket() (*s3.CreateBucketOutput, error) {
	output, err := s.S3.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(s.Config.Bucket),
	})
	if err != nil {
		return nil, err
	}

	err = s.S3.WaitUntilBucketExists(&s3.HeadBucketInput{
		Bucket: aws.String(s.Config.Bucket),
	})
	if err != nil {
		return nil, err
	}

	if config.AppConfig.AWS.S3.BucketVersioning {
		_, err = s.S3.PutBucketVersioning(&s3.PutBucketVersioningInput{
			Bucket: aws.String(s.Config.Bucket),
			VersioningConfiguration: &s3.VersioningConfiguration{
				Status: aws.String("Enabled"),
			},
		})
		if err != nil {
			return nil, err
		}
	}

	return output, nil
}

func (s *Storage) GetFiles() ([]*pb.S3Object, error) {
	output, err := s.S3.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(s.Config.Bucket),
	})
	if err != nil {
		return nil, err
	}

	objects := make([]*pb.S3Object, 0)

	for _, item := range output.Contents {
		objects = append(objects, &pb.S3Object{
			Key:          *item.Key,
			Size:         uint64(*item.Size),
			LastModified: item.LastModified.Format(time.RFC3339),
			StorageClass: *item.StorageClass,
		})
	}

	return objects, nil
}

func (s *Storage) DeleteFile(fileName string) error {
	_, err := s.S3.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(s.Config.Bucket),
		Key:    aws.String(fileName),
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) DeleteFiles() error {
	iter := s3manager.NewDeleteListIterator(s.S3, &s3.ListObjectsInput{
		Bucket: aws.String(s.Config.Bucket),
	})

	if err := s3manager.NewBatchDeleteWithClient(s.S3).Delete(aws.BackgroundContext(), iter); err != nil {
		return fmt.Errorf("unable to delete objects from bucket %s: %v", s.Config.Bucket, err)
	}

	return nil
}
