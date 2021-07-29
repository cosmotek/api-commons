package s3

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"time"

	minio "github.com/minio/minio-go/v6"
)

const defaultRegion = "us-east-1"

var ErrNotFound = errors.New("seeker can't seek")

type Bucket struct {
	client *minio.Client
	name   string
	conf   Config
}

type File interface {
	io.Reader
	io.Seeker
	io.ReaderAt
	io.Closer
}

type Config struct {
	Host, APIKey, Secret, BucketName string
	Region                           string
}

func New(conf Config) (*Bucket, error) {
	minioClient, err := minio.New(conf.Host, conf.APIKey, conf.Secret, false)
	if err != nil {
		return nil, err
	}

	ok, err := minioClient.BucketExists(conf.BucketName)
	if err != nil {
		return nil, err
	}

	if !ok {
		region := defaultRegion
		if conf.Region != "" {
			region = conf.Region
		}

		err = minioClient.MakeBucket(conf.BucketName, region)
		if err != nil {
			return nil, err
		}
	}

	return &Bucket{
		client: minioClient,
		name:   conf.BucketName,
		conf:   conf,
	}, nil
}

func (f *Bucket) Upload(filepath string, contents []byte) (string, error) {
	_, err := f.client.PutObject(
		f.name,
		filepath,
		bytes.NewBuffer(contents),
		int64(len(contents)),
		minio.PutObjectOptions{},
	)

	return fmt.Sprintf("%s/%s/%s", f.conf.Host, f.name, filepath), err
}

func (f *Bucket) GeneratePresignedURL(filepath string, filename string) (string, error) {
	url, err := f.client.PresignedGetObject(f.name, filepath, time.Hour*24*7, url.Values{
		"response-content-disposition": {fmt.Sprintf("attachment;filename=%s", filename)},
	})
	if err != nil {
		return "", err
	}

	return url.String(), nil
}

func (f *Bucket) GeneratePresignedUploadURL(ctx context.Context, filepath string, expiry time.Duration) (string, error) {
	url, err := f.client.PresignedPutObject(f.name, filepath, expiry)
	if err != nil {
		return "", err
	}

	return url.String(), nil
}

func (f *Bucket) Download(filepath string) (File, error) {
	return f.client.GetObject(f.name, filepath, minio.GetObjectOptions{})
}

func (f *Bucket) Delete(filepath string) error {
	return f.client.RemoveObject(f.name, filepath)
}

type FileInfo struct {
	URL          string
	ContentType  string
	LastModified time.Time
	Owner        string
	OwnerID      string
	Size         int64
}

func (f *Bucket) Stat(filepath string) (FileInfo, error) {
	info, err := f.client.StatObject(f.name, filepath, minio.StatObjectOptions{})
	if err != nil {
		return FileInfo{}, err
	}

	return FileInfo{
		URL:          fmt.Sprintf("%s/%s/%s", f.conf.Host, f.name, filepath),
		ContentType:  info.ContentType,
		LastModified: info.LastModified,
		Owner:        info.Owner.DisplayName,
		OwnerID:      info.Owner.ID,
		Size:         info.Size,
	}, nil
}
