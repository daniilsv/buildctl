package s3

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	aws_config "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	smithy_endpoints "github.com/aws/smithy-go/endpoints"
	"github.com/google/uuid"
)

type S3Service struct {
	client       *s3.Client
	presign      *s3.PresignClient
	bucket       string
	publicPrefix string
}

type S3Config struct {
	S3Endpoint     string
	S3Region       string
	S3PathStyle    bool
	S3Bucket       string
	S3Key          string
	S3Secret       string
	S3PublicPrefix string
}

type customResolver struct {
	config S3Config
}

func (s *customResolver) ResolveEndpoint(ctx context.Context, params s3.EndpointParameters) (
	smithy_endpoints.Endpoint, error,
) {
	u, err := url.Parse(s.config.S3Endpoint)
	if err != nil {
		return smithy_endpoints.Endpoint{}, err
	}
	u.Path = s.config.S3Bucket
	return smithy_endpoints.Endpoint{
		URI: *u,
	}, nil
}
func NewS3Service(config S3Config) *S3Service {
	creds := credentials.NewStaticCredentialsProvider(
		config.S3Key,
		config.S3Secret,
		"",
	)

	cfg, err := aws_config.LoadDefaultConfig(
		context.TODO(),
		aws_config.WithRegion(config.S3Region),
		aws_config.WithCredentialsProvider(creds),
	)

	if err != nil {
		slog.Error("S3 has been initialized!", "error", err)
	} else {
		slog.Info("S3 has been initialized!")
	}
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.EndpointResolverV2 = &customResolver{config: config}
		o.UsePathStyle = config.S3PathStyle
	})
	presign := s3.NewPresignClient(client)

	return &S3Service{
		client:       client,
		presign:      presign,
		bucket:       config.S3Bucket,
		publicPrefix: config.S3PublicPrefix,
	}
}

func (s *S3Service) PresignGet(key string) (*v4.PresignedHTTPRequest, error) {
	request, err := s.presign.PresignGetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = time.Duration(5 * int64(time.Hour))
	})
	return request, err
}

func (s *S3Service) PresignPut(folder, filename string, keepName bool) (*v4.PresignedHTTPRequest, string, error) {
	if folder == "" || filename == "" {
		return nil, "", fmt.Errorf("folder and filename are required")
	}

	newName := ""
	if keepName {
		newName = folder + "/" + filename
	} else {
		a := strings.Split(filename, ".")
		var ext = a[len(a)-1]

		newName = folder + "/" + uuid.New().String()
		if len(ext) < 5 {
			newName = newName + "." + ext
		}
	}

	request, err := s.presign.PresignPutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(newName),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = time.Duration(5 * int64(time.Hour))
	})

	return request, newName, err
}

func (s *S3Service) DeleteObject(key string) error {
	_, err := s.client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	return err
}

func (s *S3Service) HasObject(key string) (bool, error) {
	_, err := s.client.HeadObject(context.TODO(), &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s *S3Service) GetPublicURL(key string) string {
	if key == "" {
		return ""
	}
	return s.publicPrefix + "/" + key
}
