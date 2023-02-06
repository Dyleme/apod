package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var ErrBucketNotExist = errors.New("bucket not exist")

// MinioStorage is a struct that provides methods to store files in minio storage.
type MinioStorage struct {
	client   minio.Client
	endpoint string
}

// MinioConnfig is a config to make connection with minio storage.
type Config struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
}

func GetConfig() (*Config, error) {
	mnhost := os.Getenv("MNHOST")
	mnport := os.Getenv("MNPORT")
	accessKey := os.Getenv("MNACCESSKEYID")
	secretKey := os.Getenv("MNSECRETACCESSKEY")
	ssl := os.Getenv("MNUSESSL")

	useSsl, err := strconv.ParseBool(ssl)
	if err != nil {
		return nil, fmt.Errorf("cant parse %q into bool: %w", ssl, err)
	}

	return &Config{
		Endpoint:        mnhost + ":" + mnport,
		AccessKeyID:     accessKey,
		SecretAccessKey: secretKey,
		UseSSL:          useSsl,
	}, nil
}

// NewMinoStorage is a constructor to the MinioStoage.
// Returns error if the connection is denied.
func NewMinioStorage(cfg Config) (*MinioStorage, error) {
	cl, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("can not initialize storage: %w", err)
	}

	return &MinioStorage{client: *cl, endpoint: cfg.Endpoint}, nil
}

func (m *MinioStorage) GetURL(bucket, filename string) string {
	return "http://" + m.endpoint + "/" + bucket + "/" + filename
}

// UploadFile method upload provided file to the minio storage and returns path to the file.
func (m *MinioStorage) UploadFile(ctx context.Context, bucket, filename string, data []byte) (string, error) {
	exist, err := m.client.BucketExists(ctx, bucket)
	if err != nil {
		return "", fmt.Errorf("check bucket existing: %w", err)
	}

	if !exist {
		err := m.createPublicBucket(ctx, bucket)
		if err != nil {
			return "", err
		}
	}

	bf := bytes.NewBuffer(data)

	_, err = m.client.PutObject(ctx, bucket, filename, bf, int64(bf.Len()), minio.PutObjectOptions{})

	if err != nil {
		return "", fmt.Errorf("can not upload file: %w", err)
	}

	return m.GetURL(bucket, filename), nil
}

func (m *MinioStorage) createPublicBucket(ctx context.Context, bucket string) error {
	err := m.client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
	if err != nil {
		return fmt.Errorf("make bucket: %w", err)
	}

	const publicPolicy = `{"Version": "2012-10-17","Statement": [{"Action": ["s3:GetObject"],"Effect": "Allow","Principal": {"AWS": ["*"]},"Resource": ["arn:aws:s3:::%s/*"],"Sid": ""}]}`

	err = m.client.SetBucketPolicy(ctx, bucket, fmt.Sprintf(publicPolicy, bucket))
	if err != nil {
		return fmt.Errorf("set bucket policy: %w", err)
	}

	return nil
}
