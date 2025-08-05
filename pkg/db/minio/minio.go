package minio_db

import (
	"caviar/internal/config"
	"context"
	"io"
	"log"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/pkg/errors"
)

type PutObjectParams struct {
	ObjectName  string
	Reader      io.Reader
	Size        int64
	ContentType string
}

type GetObjectURLParams struct {
	ObjectName string
	Expiry     time.Duration
	ReqParams  url.Values
}

type Minio struct {
	Client     *minio.Client
	BucketName string
}

func NewMinio(ctx context.Context, cfg config.Minio) (*Minio, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize MinIO client")
	}

	m := &Minio{Client: client, BucketName: cfg.BucketName}

	exists, err := client.BucketExists(ctx, m.BucketName)
	if err != nil {
		return nil, errors.Wrapf(err, "could not check if bucket %q exists", m.BucketName)
	}
	if !exists {
		if err := client.MakeBucket(ctx, m.BucketName, minio.MakeBucketOptions{}); err != nil {
			return nil, errors.Wrapf(err, "failed to create bucket %q", m.BucketName)
		}
		log.Printf("[MinIO] created bucket %s", m.BucketName)
	}

	return m, nil
}

func (m *Minio) Close() error {
	return nil
}


func (m *Minio) PutObject(ctx context.Context, params PutObjectParams) (minio.UploadInfo, error) {
	info, err := m.Client.PutObject(
		ctx,
		m.BucketName,
		params.ObjectName,
		params.Reader,
		params.Size,
		minio.PutObjectOptions{ContentType: params.ContentType},
	)
	if err != nil {
		return minio.UploadInfo{}, errors.Wrapf(err, "failed to upload object %q to bucket %q", params.ObjectName, m.BucketName)
	}
	return info, nil
}

func (m *Minio) GetObject(
	ctx context.Context,
	objectName string,
	opts minio.GetObjectOptions,
) (*minio.Object, error) {
	obj, err := m.Client.GetObject(
		ctx,
		m.BucketName,
		objectName,
		opts,
	)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get object %q from bucket %q", objectName, m.BucketName)
	}
	return obj, nil
}

func (m *Minio) RemoveObject(
	ctx context.Context,
	objectName string,
	opts minio.RemoveObjectOptions,
) error {
	if err := m.Client.RemoveObject(
		ctx,
		m.BucketName,
		objectName,
		opts,
	); err != nil {
		return errors.Wrapf(err, "failed to remove object %q from bucket %q", objectName, m.BucketName)
	}
	return nil
}

func (m *Minio) GetObjectURL(ctx context.Context, params GetObjectURLParams) (string, error) {
	u, err := m.Client.PresignedGetObject(
		ctx,
		m.BucketName,
		params.ObjectName,
		params.Expiry,
		params.ReqParams,
	)
	if err != nil {
		return "", errors.Wrapf(err, "failed to generate presigned URL for object %q", params.ObjectName)
	}
	return u.String(), nil
}
