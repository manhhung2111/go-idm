package file

import (
	"context"
	"io"

	"github.com/manhhung2111/go-idm/internal/config"
	"github.com/manhhung2111/go-idm/internal/utils"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type S3Client struct {
	minioClient *minio.Client
	bucket      string
	logger      *zap.Logger
}

func NewS3Client(downloadConfig config.Download, logger *zap.Logger) (Client, error) {
	minioClient, err := minio.New(downloadConfig.Address, &minio.Options{
		Creds:  credentials.NewStaticV4(downloadConfig.Username, downloadConfig.Password, ""),
		Secure: false,
	})

	if err != nil {
		logger.With(zap.Error(err)).Error("failed to create minio client")
		return nil, err
	}

	return &S3Client{
		minioClient: minioClient,
		bucket:      downloadConfig.Bucket,
		logger:      logger,
	}, nil
}

func (s S3Client) Read(ctx context.Context, filePath string) (io.ReadCloser, error) {
	logger := utils.LoggerWithContext(ctx, s.logger).With(zap.String("file_path", filePath))

	obj, err := s.minioClient.GetObject(
		ctx,
		s.bucket,
		filePath,
		minio.GetObjectOptions{},
	)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to get s3 object")
		return nil, status.Error(codes.Internal, "failed to get s3 object")
	}

	return obj, nil
}

func (s S3Client) Write(ctx context.Context, filePath string) (io.WriteCloser, error) {
	logger := utils.LoggerWithContext(ctx, s.logger).With(zap.String("file_path", filePath))

	pr, pw := io.Pipe()

	go func() {
		defer pr.Close()

		_, err := s.minioClient.PutObject(
			ctx,
			s.bucket,
			filePath,
			pr,
			-1, // unknown size (streaming)
			minio.PutObjectOptions{
				ContentType: "application/octet-stream",
				NumThreads: 8,             
				ConcurrentStreamParts: true, 
			},
		)

		if err != nil {
			logger.With(zap.Error(err)).Error("failed to upload to s3")
			_ = pw.CloseWithError(err)
			return
		}

		_ = pw.Close()
	}()

	return pw, nil
}
