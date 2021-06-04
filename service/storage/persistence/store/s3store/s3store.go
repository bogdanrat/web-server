package s3store

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	pb "github.com/bogdanrat/web-server/contracts/proto/storage_service"
	"github.com/bogdanrat/web-server/service/storage/config"
	"github.com/bogdanrat/web-server/service/storage/persistence/store"
	"github.com/rlmcpherson/s3gof3r"
	"io"
	"log"
	"sync"
	"time"
)

type S3Store struct {
	S3           *s3.S3
	S3Gofer      *s3gof3r.S3 // provides fast, parallelized, streaming access to Amazon S3
	Bucket       *s3gof3r.Bucket
	Config       *s3gof3r.Config
	UploaderPool sync.Pool
	Session      *session.Session
}

func New(sess *session.Session, s3Config config.S3Config) store.Store {
	stsCredentials := stscreds.NewCredentials(sess, config.AppConfig.AWS.AssumeRole, func(provider *stscreds.AssumeRoleProvider) {
		provider.ExternalID = &config.AppConfig.AWS.ExternalID
	})
	sess.Config.Credentials = stsCredentials

	storage := &S3Store{}
	storage.Session = sess

	storage.S3 = s3.New(sess, &aws.Config{
		Credentials: stsCredentials,
	})
	creds, _ := stsCredentials.Get()

	storage.S3Gofer = s3gof3r.New(s3Config.Domain, s3gof3r.Keys{
		AccessKey: creds.AccessKeyID,
		SecretKey: creds.SecretAccessKey,
	})

	storage.Bucket = storage.S3Gofer.Bucket(s3Config.Bucket)

	storage.Config = s3gof3r.DefaultConfig
	storage.Config.Concurrency = s3Config.Concurrency                             // number of parts to get or put concurrently
	storage.Config.PartSize = int64(s3Config.PartSize)                            // initial part size in bytes to use for multipart gets or put
	storage.Config.NTry = s3Config.MaxAttempts                                    // maximum attempts for each part
	storage.Config.Client.Timeout = time.Second * time.Duration(s3Config.Timeout) // includes connection time, any redirects, and reading the response body

	storage.UploaderPool = sync.Pool{
		New: func() interface{} {
			uploader := s3manager.NewUploader(storage.Session)
			// The number of goroutines to spin up in parallel per call to Upload when sending parts
			uploader.Concurrency = s3Config.Concurrency
			return uploader
		},
	}

	return storage
}

func (s *S3Store) Init() error {
	_, err := s.S3.HeadBucket(
		&s3.HeadBucketInput{
			Bucket: aws.String(s.Bucket.Name),
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
				log.Printf("S3 Bucket created at location: %s\n", *output.Location)
			}
		} else {
			return err
		}
	} else {
		log.Println("S3 Bucket already exists. Skipping bucket creation.")
	}

	log.Printf("Initialized S3 Storage Engine in %s\n", s.Bucket.Name)
	return nil
}

func (s *S3Store) createBucket() (*s3.CreateBucketOutput, error) {
	output, err := s.S3.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(s.Bucket.Name),
	})
	if err != nil {
		return nil, err
	}

	err = s.S3.WaitUntilBucketExists(&s3.HeadBucketInput{
		Bucket: aws.String(s.Bucket.Name),
	})
	if err != nil {
		return nil, err
	}

	if config.AppConfig.AWS.S3.BucketVersioning {
		_, err = s.S3.PutBucketVersioning(&s3.PutBucketVersioningInput{
			Bucket: aws.String(s.Bucket.Name),
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

func (s *S3Store) Put(key string, body io.Reader) error {
	// Get() first checks if there are any available instances within the pool to return. If not, calls New() to create a new one.
	uploader := s.UploaderPool.Get().(*s3manager.Uploader)
	// Put(): place the instance back in the pool for use by other processes.
	defer s.UploaderPool.Put(uploader)

	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s.Bucket.Name),
		Key:    aws.String(key),
		Body:   body,
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *S3Store) Get(key string, writer io.Writer) error {
	// GetReader() provides a reader and downloads data using parallel ranged get requests.
	// Data from the requests are ordered and written sequentially.
	reader, _, err := s.Bucket.GetReader(key, s.Config)
	if err != nil {
		return err
	}
	defer reader.Close()

	if _, err = io.Copy(writer, reader); err != nil {
		return err
	}
	return nil
}

func (s *S3Store) GetAll() ([]*pb.StorageObject, error) {
	output, err := s.S3.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(s.Bucket.Name),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			code := aerr.Code()
			switch code {
			case "AccessDenied":
				s.retrieveNewCredentials()
				return s.GetAll()
			}
		} else {
			return nil, err
		}

		return nil, err
	}

	objects := make([]*pb.StorageObject, 0)

	for _, item := range output.Contents {
		objects = append(objects, &pb.StorageObject{
			Key:          *item.Key,
			Size:         uint64(*item.Size),
			LastModified: item.LastModified.Format(time.RFC3339),
			StorageClass: *item.StorageClass,
		})
	}

	return objects, nil
}

func (s *S3Store) Delete(fileName string) error {
	if err := s.Bucket.Delete(fileName); err != nil {
		return err
	}
	return nil
}

func (s *S3Store) DeleteAll(prefix ...string) error {
	input := &s3.ListObjectsInput{
		Bucket: aws.String(s.Bucket.Name),
	}
	if len(prefix) == 1 {
		input.Prefix = aws.String(prefix[0])
	}

	iter := s3manager.NewDeleteListIterator(s.S3, input)

	if err := s3manager.NewBatchDeleteWithClient(s.S3).Delete(aws.BackgroundContext(), iter); err != nil {
		return fmt.Errorf("unable to delete objects from bucket %s: %v", s.Bucket.Name, err)
	}

	return nil
}

func (s *S3Store) retrieveNewCredentials() {
	stsCredentials := stscreds.NewCredentials(s.Session, config.AppConfig.AWS.AssumeRole, func(provider *stscreds.AssumeRoleProvider) {
		provider.ExternalID = &config.AppConfig.AWS.ExternalID
	})
	s.Session.Config.Credentials = stsCredentials
	s.S3 = s3.New(s.Session, &aws.Config{
		Credentials: stsCredentials,
	})
}
