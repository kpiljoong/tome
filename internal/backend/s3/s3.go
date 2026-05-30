package s3

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"github.com/kpiljoong/tome/internal/logx"
	"github.com/kpiljoong/tome/internal/model"
	"github.com/kpiljoong/tome/internal/paths"
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

	// Check if object already exists in S3
	s3Key := filepath.ToSlash(filepath.Join(b.Prefix, remotePath))
	exists, err := b.Exists(s3Key)
	if err != nil {
		return fmt.Errorf("existence check failed: %w", err)
	}
	if exists {
		// Compare local and remote hashes (via HeadObject ETag, if enabled)
		// — but S3 ETag isn't always reliable.
		// So we'll just log that we are skipping based on existence.
		logx.Info("☁️  Skipped (already exists): s3://%s/%s", b.Bucket, s3Key)
		return nil
	}

	// Upload if not exists
	_, err = b.Uploader.Upload(context.TODO(), &s3.PutObjectInput{
		Bucket:       aws.String(b.Bucket),
		Key:          aws.String(s3Key),
		Body:         file,
		StorageClass: types.StorageClassOnezoneIa,
	})
	if err != nil {
		return fmt.Errorf("failed to upload file %s to s3://%s/%s: %w", localPath, b.Bucket, s3Key, err)
	}

	logx.Success("⬆️  %s → s3://%s/%s", localPath, b.Bucket, s3Key)
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
	var failures []error
	query = strings.ToLower(query)
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
			key := aws.ToString(obj.Key)
			if !strings.HasSuffix(key, ".json") {
				continue
			}

			getOut, err := b.Client.GetObject(context.TODO(), &s3.GetObjectInput{
				Bucket: aws.String(b.Bucket),
				Key:    aws.String(key),
			})
			if err != nil {
				failures = append(failures, fmt.Errorf("get journal object %s: %w", key, err))
				continue
			}

			entry, err := decodeJournalObject(key, getOut.Body)
			if err != nil {
				failures = append(failures, err)
				continue
			}
			if journalEntryMatchesQuery(entry, query) {
				entries = append(entries, entry)
			}
		}
	}
	if len(failures) > 0 {
		return nil, fmt.Errorf("list s3 journal %s completed with %d failure(s): %w", namespace, len(failures), errors.Join(failures...))
	}

	// Sort by newest first
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Timestamp.After(entries[j].Timestamp)
	})
	return entries, nil
}

func (b *S3Backend) GetBlobByHash(hash string) ([]byte, error) {
	key := filepath.ToSlash(filepath.Join(b.Prefix, paths.RemoteBlobPath(hash)))
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

func (b *S3Backend) GeneratePresignedURL(key string, expiry time.Duration) (string, error) {
	presigner := s3.NewPresignClient(b.Client)

	out, err := presigner.PresignGetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(b.Bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expiry))
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return out.URL, nil
}

func (b *S3Backend) BlobKey(hash string) string {
	return filepath.ToSlash(filepath.Join(b.Prefix, "blobs", paths.SanitizeHash(hash)))
}

func decodeJournalObject(key string, body io.ReadCloser) (*model.JournalEntry, error) {
	defer body.Close()

	var entry model.JournalEntry
	if err := json.NewDecoder(body).Decode(&entry); err != nil {
		return nil, fmt.Errorf("decode journal object %s: %w", key, err)
	}
	return &entry, nil
}

func journalEntryMatchesQuery(entry *model.JournalEntry, query string) bool {
	return query == "" ||
		strings.Contains(strings.ToLower(entry.Filename), query) ||
		strings.Contains(strings.ToLower(entry.FullPath), query)
}
