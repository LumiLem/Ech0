package util

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
	"time"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type minioStorage struct {
	client     *minio.Client
	bucketName string
}

func NewMinioStorage(
	endpoint, accessKey, secretKey, bucketName, region, provider string,
	secure bool,
) (ObjectStorage, error) {
	if accessKey == "" || secretKey == "" || bucketName == "" {
		return nil, fmt.Errorf("invalid S3 configuration")
	}

	normalizedProvider := strings.ToLower(strings.TrimSpace(provider))

	// 根据 Provider 自动调整 endpoint 和 SSL
	switch normalizedProvider {
	case string(commonModel.R2):
		// Cloudflare R2 强制 HTTPS
		secure = true
		if region == "" {
			region = "auto"
		}
	case string(commonModel.AWS):
		// AWS 可以使用 region 拼接 endpoint
		if region != "" && endpoint == "" {
			endpoint = fmt.Sprintf("s3.%s.amazonaws.com", region)
		}
	case string(commonModel.MINIO):
		// MinIO region 可空
	case string(commonModel.OTHER), string(commonModel.ALIYUN), string(commonModel.TENCENT):
		if region == "" {
			region = "auto"
		}
	case "":
		if region == "" {
			region = "auto"
		}
	default:
		if region == "" {
			region = "auto"
		}
	}

	if endpoint == "" {
		return nil, fmt.Errorf("invalid S3 configuration")
	}

	client, err := minio.New(endpoint, &minio.Options{
		Creds:              credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure:             secure,
		Region:             region,
		BucketLookupViaURL: CustomBucketLookupViaURL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Minio client: %w", err)
	}

	// 检查桶是否存在
	exists, err := client.BucketExists(context.Background(), bucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to check if bucket exists: %w", err)
	}
	if !exists {
		if err := client.MakeBucket(context.Background(), bucketName, minio.MakeBucketOptions{Region: region}); err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	return &minioStorage{client: client, bucketName: bucketName}, nil
}

// DeleteObject implements storage.ObjectStorage.
func (m *minioStorage) DeleteObject(ctx context.Context, objectName string) error {
	// Delete an object from the bucket
	err := m.client.RemoveObject(ctx, m.bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return err
	}
	return nil
}

// Download implements storage.ObjectStorage.
func (m *minioStorage) Download(ctx context.Context, objectName string) (io.ReadCloser, error) {
	// Download an object from the bucket
	obj, err := m.client.GetObject(ctx, m.bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	return obj, nil
}

// ListObjects implements storage.ObjectStorage.
func (m *minioStorage) ListObjects(ctx context.Context, prefix string) ([]string, error) {
	// List objects in the bucket with the specified prefix
	objectCh := m.client.ListObjects(ctx, m.bucketName, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})

	var objects []string
	for obj := range objectCh {
		if obj.Err != nil {
			return nil, obj.Err
		}
		objects = append(objects, obj.Key)
	}
	return objects, nil
}

// ListObjectStream streams object names to a channel to avoid high memory usage.
func (m *minioStorage) ListObjectStream(ctx context.Context, prefix string) (<-chan string, error) {
	objectCh := m.client.ListObjects(ctx, m.bucketName, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})

	resultCh := make(chan string)

	go func() {
		defer close(resultCh)
		for obj := range objectCh {
			if obj.Err != nil {
				// Handle error
				return
			}
			select {
			case resultCh <- obj.Key:
			case <-ctx.Done():
				return
			}
		}
	}()

	return resultCh, nil
}

// PresignURL implements storage.ObjectStorage.
func (m *minioStorage) PresignURL(
	ctx context.Context,
	objectName string,
	expiry time.Duration,
	method string,
) (string, error) {
	switch method {
	case "GET":
		// 下载类型预签名 URL
		reqParams := make(url.Values)
		reqParams.Set("response-content-disposition", "attachment")
		presignedURL, err := m.client.PresignedGetObject(
			ctx,
			m.bucketName,
			objectName,
			expiry,
			reqParams,
		)
		if err != nil {
			return "", err
		}
		return presignedURL.String(), nil

	case "PUT":
		// 上传类型预签名 URL
		presignedURL, err := m.client.PresignedPutObject(ctx, m.bucketName, objectName, expiry)
		if err != nil {
			return "", err
		}
		return presignedURL.String(), nil

	default:
		return "", fmt.Errorf("unsupported method: %s, must be GET or PUT", method)
	}
}

// Upload implements storage.ObjectStorage.
func (m *minioStorage) Upload(
	ctx context.Context,
	objectName string,
	r io.Reader,
	contentType string,
) error {
	var objectSize int64 = -1
	// Try to determine the size of the reader
	if s, ok := r.(interface {
		Size() int64
	}); ok {
		objectSize = s.Size()
	} else if f, ok := r.(*os.File); ok {
		if stat, err := f.Stat(); err == nil {
			objectSize = stat.Size()
		}
	}

	_, err := m.client.PutObject(
		ctx,
		m.bucketName,
		objectName,
		r,
		objectSize,
		minio.PutObjectOptions{
			ContentType: contentType,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to upload object: %w", err)
	}
	return nil
}

// IsTencentCOSEndpoint - Match if it is exactly Tencent Cloud COS endpoint.
func IsTencentCOSEndpoint(endpointURL url.URL) bool {
	hostname := endpointURL.Hostname()

	return strings.HasSuffix(hostname, "myqcloud.com") ||
		strings.HasSuffix(hostname, "tencentcos.cn")
}

// CustomBucketLookupViaURL implements minio.BucketLookupViaURL function.
func CustomBucketLookupViaURL(endpointURL url.URL, bucketName string) minio.BucketLookupType {
	// 腾讯云COS支持VirtualHostStyle
	if IsTencentCOSEndpoint(endpointURL) {
		return minio.BucketLookupDNS
	}
	return minio.BucketLookupAuto
}
