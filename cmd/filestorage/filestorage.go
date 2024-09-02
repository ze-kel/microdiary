package storage

import (
	"bytes"
	"context"
	"io"
	"log"
	"os"
	"path"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type GenericStorage interface {
	SaveFile(ctx context.Context, r io.Reader, fileName string, mimeType string) error
	ReadFile(ctx context.Context, fileName string) (file []byte, err error)
}

type FileStorage struct {
	basePath string
}

func NewLocal(basePath string) *FileStorage {
	err := os.MkdirAll(basePath, os.ModePerm)
	if err != nil {
		panic("failed to make files directory")
	}

	return &FileStorage{
		basePath: basePath,
	}
}

func (f *FileStorage) SaveFile(ctx context.Context, r io.Reader, fileName string, mimeType string) error {
	out, err := os.Create(path.Join(f.basePath, fileName))

	if err != nil {
		return err
	}

	defer out.Close()

	_, err = io.Copy(out, r)

	if err != nil {
		return err
	}

	return nil
}

func (f *FileStorage) ReadFile(ctx context.Context, fileName string) (file []byte, err error) {
	return os.ReadFile(path.Join(f.basePath, fileName))
}

type S3Storage struct {
	minioClient *minio.Client
	bucketName  string
}

type S3Params struct {
	Url        string
	ID         string
	Secret     string
	BucketName string
}

func NewS3(ctx context.Context, p S3Params) *S3Storage {

	c := credentials.NewStaticV4(p.ID, p.Secret, "")

	minioClient, err := minio.New(p.Url, &minio.Options{
		Creds:  c,
		Secure: true,
	})
	if err != nil {
		log.Fatalln(err)
	}

	err = minioClient.MakeBucket(ctx, p.BucketName, minio.MakeBucketOptions{})
	if err != nil {
		// Check to see if we already own this bucket (which happens if you run this twice)
		exists, errBucketExists := minioClient.BucketExists(ctx, p.BucketName)
		if errBucketExists == nil && exists {
			log.Printf("We already own %s\n", p.BucketName)
		} else {
			log.Fatalln(err)
		}
	} else {
		log.Printf("Successfully created %s\n", p.BucketName)
	}

	return &S3Storage{
		minioClient: minioClient,
		bucketName:  p.BucketName,
	}
}

func (f *S3Storage) SaveFile(ctx context.Context, r io.Reader, fileName string, mimeType string) error {

	buf := new(bytes.Buffer)
	buf.ReadFrom(r)

	_, err := f.minioClient.PutObject(ctx, f.bucketName, fileName, buf, int64(buf.Len()), minio.PutObjectOptions{ContentType: mimeType})

	return err
}

func (f *S3Storage) ReadFile(ctx context.Context, fileName string) (file []byte, err error) {

	v, err := f.minioClient.GetObject(ctx, f.bucketName, fileName, minio.GetObjectOptions{})

	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(v)

	return buf.Bytes(), err
}
