package service

import (
	"bytes"
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type S3Service struct {
    Client *s3.S3
    Bucket string
}

func NewS3Service(bucket string) *S3Service {
    sess := session.Must(session.NewSession())
    return &S3Service{
        Client: s3.New(sess),
        Bucket: bucket,
    }
}

func (s *S3Service) UploadFile(ctx context.Context, key string, body []byte) error {
    _, err := s.Client.PutObject(&s3.PutObjectInput{
        Bucket: aws.String(s.Bucket),
        Key:    aws.String(key),
        Body:   aws.ReadSeekCloser(bytes.NewReader(body)),
    })
    if err != nil {
        return fmt.Errorf("failed to upload file: %w", err)
    }
    return nil
}
