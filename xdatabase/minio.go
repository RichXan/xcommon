package xdatabase

import (
	"bytes"
	"context"
	"io"
	"mime"
	"os"
	"path/filepath"
	"time"

	"github.com/richxan/xpkg/xlog"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var defaultContentType = "application/octet-stream"

type MinioConfig struct {
	Endpoint  string `yaml:"endpoint"`
	Bucket    string `yaml:"bucket"`
	AccessKey string `yaml:"access_key"`
	SecretKey string `yaml:"secret_key"`
}

type Minio struct {
	Conf   MinioConfig
	Client *minio.Client
	logger *xlog.Logger
}

func (m *Minio) NewClient(log *xlog.Logger) error {
	client, err := minio.New(m.Conf.Endpoint, &minio.Options{
		Creds: credentials.NewStaticV4(m.Conf.AccessKey, m.Conf.SecretKey, ""),
	})
	if err != nil {
		return err
	}
	m.Client = client
	m.logger = log
	return nil
}

// UploadFileByPath 根据文件完整路径上传文件
func (m *Minio) UploadFileByPath(objectName, filePath string, opts minio.PutObjectOptions) error {
	_, err := m.Client.FPutObject(context.Background(), m.Conf.Bucket, objectName, filePath, opts)
	if err != nil {
		m.logger.Error().Err(err).Msgf("UploadFileByPath error: %s", err)
	}
	return err
}

// UploadFileByBytes 根据文件字节流上传文件
func (m *Minio) UploadFileByBytes(objectName string, fileBytes []byte, opts minio.PutObjectOptions) error {
	_, err := m.Client.PutObject(context.Background(), m.Conf.Bucket, objectName, bytes.NewReader(fileBytes), int64(len(fileBytes)), opts)
	if err != nil {
		m.logger.Error().Err(err).Msgf("UploadFileByBytes error: %s", err)
	}
	return err
}

// UploadFileByBuffer 根据文件缓冲区上传文件
func (m *Minio) UploadFileByBuffer(objectName string, buf *bytes.Buffer, opts minio.PutObjectOptions) error {
	_, err := m.Client.PutObject(context.Background(), m.Conf.Bucket, objectName, buf, int64(buf.Len()), opts)
	if err != nil {
		m.logger.Error().Err(err).Msgf("UploadFileByBuffer error: %s", err)
	}
	return err
}

// UploadFile 根据文件流上传文件
func (m *Minio) UploadFile(objectName string, fileReader *os.File, opts minio.PutObjectOptions) error {
	fileStat, err := fileReader.Stat()
	if err != nil {
		m.logger.Error().Err(err).Msgf("UploadFile error: %s", err)
		return err
	}
	fileSize := fileStat.Size()
	if opts.ContentType == "" {
		if opts.ContentType = mime.TypeByExtension(filepath.Ext(fileStat.Name())); opts.ContentType == "" {
			opts.ContentType = defaultContentType
		}
	}
	_, err = m.Client.PutObject(context.Background(), m.Conf.Bucket, objectName, fileReader, fileSize, opts)
	if err != nil {
		m.logger.Error().Err(err).Msgf("UploadFile error: %s", err)
	}
	return err
}

// GetFileBytes 获取文件字节流
func (m *Minio) GetFileBytes(objectName string, opts minio.GetObjectOptions) ([]byte, error) {
	object, err := m.Client.GetObject(context.Background(), m.Conf.Bucket, objectName, opts)
	if err != nil {
		m.logger.Error().Err(err).Msgf("GetFileBytes error: %s", err)
		return nil, err
	}
	defer object.Close()
	var fileBytes []byte
	// 逐块读取对象内容
	bufferSize := 1024 * 10 // 10 kb 缓冲区大小
	buffer := make([]byte, 0, bufferSize)
	for {
		n, rErr := object.Read(buffer)
		if rErr == io.EOF {
			break
		}
		if rErr != nil {
			m.logger.Error().Err(rErr).Msgf("GetFileBytes error: %s", rErr)
			return nil, rErr
		}
		// 追加读取的数据
		fileBytes = append(fileBytes, buffer[:n]...)
		buffer = buffer[:0]
	}
	return fileBytes, err
}

// GetFileReader 获取文件流
func (m *Minio) GetFileReader(objectName string, opts minio.GetObjectOptions) (io.Reader, error) {
	object, err := m.Client.GetObject(context.Background(), m.Conf.Bucket, objectName, opts)
	if err != nil {
		m.logger.Error().Err(err).Msgf("GetFileReader error: %s", err)
		return nil, err
	}
	return object, err
}

func (m *Minio) GetFileUrl(objectName string, timeOut *time.Duration) string {
	if timeOut == nil {
		// 默认24小时有效
		t := time.Hour * 24
		timeOut = &t
	}
	// 生成预签名的 URL
	resignedURL, err := m.Client.PresignedGetObject(context.Background(), m.Conf.Bucket, objectName, *timeOut, nil)
	if err != nil || resignedURL == nil {
		m.logger.Error().Err(err).Msgf("GetFileUrl error: %s", err)
		return ""
	}
	return resignedURL.String()
}
