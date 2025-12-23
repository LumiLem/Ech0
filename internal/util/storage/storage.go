package util

import (
	"context"
	"errors"
	"io"
	"mime/multipart"
	"os"
	"time"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
)

// UploadFile 根据文件类型和存储类型上传文件
func UploadFile(
	file *multipart.FileHeader,
	fileType commonModel.UploadFileType,
	storageType commonModel.FileStorageType,
	userID uint,
) (string, error) {
	if file == nil {
		return "", errors.New(commonModel.NO_FILE_UPLOAD_ERROR)
	}

	switch storageType {
	case commonModel.LOCAL_FILE:
		return UploadFileToLocal(file, fileType, userID)
	case commonModel.S3_FILE:
		// TODO: Implement S3 file upload
	default:
		return "", errors.New(commonModel.NO_FILE_STORAGE_ERROR)
	}

	return "", errors.New(commonModel.NO_FILE_STORAGE_ERROR)
}

// IsAllowedType 检查Content-Type是否在允许的类型列表中
func IsAllowedType(contentType string, allowedTypes []string) bool {
	for _, allowed := range allowedTypes {
		if contentType == allowed {
			return true
		}
	}
	return false
}

// createDirIfNotExist 创建目录如果不存在
func createDirIfNotExist(imagePath string) error {
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		if err := os.MkdirAll(imagePath, os.ModePerm); err != nil {
			return err
		}
	}
	return nil
}

// FileExists 文件是否存在
func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

// ObjectStorage 对象存储接口
type ObjectStorage interface {
	// Upload 上传文件到对象存储
	Upload(ctx context.Context, objectName string, r io.Reader, contentType string) error

	// Download 下载对象存储中的文件
	Download(ctx context.Context, objectName string) (io.ReadCloser, error)

	// ListObjects 列出对象存储中的文件
	ListObjects(ctx context.Context, prefix string) ([]string, error)

	// ListObjectStream 列出对象存储中的文件流
	ListObjectStream(ctx context.Context, prefix string) (<-chan string, error)

	// DeleteObject 删除对象存储中的文件
	DeleteObject(ctx context.Context, objectName string) error

	// PresignURL 生成对象存储中文件的临时访问链接
	PresignURL(
		ctx context.Context,
		objectName string,
		expiry time.Duration,
		method string,
	) (string, error)
}
