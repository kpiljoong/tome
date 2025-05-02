package s3

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"github.com/kpiljoong/tome/pkg/logx"
	"github.com/kpiljoong/tome/pkg/model"
)

type S3Backend struct {
	Client   *s3.Client
	Uploader *manager.Uploader
	Bucket   string
	Prefix   string
}

func NewS3Backend(bucket, prefix string) (*S3Backend, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(cfg)
	uploader := manager.NewUploader(client)

	return &S3Backend{
		Client:   s3.NewFromConfig(cfg),
		Uploader: uploader,
		Bucket:   bucket,
		Prefix:   strings.TrimSuffix(prefix, "/"),
	}, nil
}

func (b *S3Backend) UploadDir(localRoot, remotePrefix string) error {
	return filepath.Walk(localRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(localRoot, path)
		if err != nil {
			return err
		}

		s3Key := filepath.ToSlash(filepath.Join(remotePrefix, relPath))
		return b.UploadFile(path, s3Key)
	})
}

func (b *S3Backend) UploadFile(localPath, remotePath string) error {
	file, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", localPath, err)
	}
	defer file.Close()

	s3Key := filepath.ToSlash(filepath.Join(b.Prefix, remotePath))

	_, err = b.Uploader.Upload(context.TODO(), &s3.PutObjectInput{
		Bucket:       &b.Bucket,
		Key:          aws.String(s3Key),
		Body:         file,
		StorageClass: types.StorageClassOnezoneIa,
	})
	if err != nil {
		return fmt.Errorf("failed to upload file %s to s3://%s/%s: %w", localPath, b.Bucket, s3Key, err)
	}

	logx.Success("⬆️  %s → s3://%s/%s", localPath, b.Bucket, s3Key)
	// fmt.Printf(" Uploaded %s to s3://%s/%s\n", localPath, b.Bucket, s3Key)
	return nil
}

func (b *S3Backend) Exists(remotePath string) (bool, error) {
	_, err := b.Client.HeadObject(context.TODO(), &s3.HeadObjectInput{
		Bucket: &b.Bucket,
		Key:    &remotePath,
	})
	if err != nil {
		if strings.Contains(err.Error(), "NotFound") {
			return false, nil
		}
		return false, fmt.Errorf("failed to check existence of s3://%s/%s: %w", b.Bucket, remotePath, err)
	}

	return true, nil
}

func (b *S3Backend) ListJournal(namespace, query string) ([]*model.JournalEntry, error) {
	prefix := filepath.ToSlash(filepath.Join(b.Prefix, "journals", namespace)) + "/"

	var entries []*model.JournalEntry
	paginator := s3.NewListObjectsV2Paginator(b.Client, &s3.ListObjectsV2Input{
		Bucket: aws.String(b.Bucket),
		Prefix: aws.String(prefix),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return nil, fmt.Errorf("error listing jorunal: %w", err)
		}

		for _, obj := range page.Contents {
			if !strings.HasSuffix(*obj.Key, ".json") {
				continue
			}

			getOut, err := b.Client.GetObject(context.TODO(), &s3.GetObjectInput{
				Bucket: aws.String(b.Bucket),
				Key:    obj.Key,
			})
			if err != nil {
				continue
			}
			defer getOut.Body.Close()

			var entry model.JournalEntry
			if err := json.NewDecoder(getOut.Body).Decode(&entry); err != nil {
				continue
			}

			if strings.Contains(strings.ToLower(entry.Filename), strings.ToLower(query)) ||
				strings.Contains(strings.ToLower(entry.FullPath), strings.ToLower(query)) {
				entries = append(entries, &entry)
			}
		}
	}
	return entries, nil
}

func (b *S3Backend) GetBlobByHash(hash string) ([]byte, error) {
	safeHash := strings.ReplaceAll(hash, ":", "-")
	key := filepath.ToSlash(filepath.Join(b.Prefix, "blobs", safeHash))
	resp, err := b.Client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(b.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch blob: %w", err)
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func (b *S3Backend) ListNamespaces() ([]string, error) {
	prefix := filepath.ToSlash(filepath.Join(b.Prefix, "journals")) + "/"

	paginator := s3.NewListObjectsV2Paginator(b.Client, &s3.ListObjectsV2Input{
		Bucket:    aws.String(b.Bucket),
		Prefix:    aws.String(prefix),
		Delimiter: aws.String("/"),
	})

	var namespaces []string
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return nil, err
		}
		for _, p := range page.CommonPrefixes {
			ns := strings.TrimPrefix(*p.Prefix, prefix)
			ns = strings.TrimSuffix(ns, "/")
			namespaces = append(namespaces, ns)
		}
	}
	return namespaces, nil
}

func (b *S3Backend) Describe() string {
	if b.Prefix != "" {
		return fmt.Sprintf("s3://%s/%s", b.Bucket, b.Prefix)
	}
	return fmt.Sprintf("s3://%s", b.Bucket)
}
