package s3

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/kpiljoong/tome/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListJournalReturnsErrorForDecodeFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Query().Get("list-type") == "2":
			w.Header().Set("Content-Type", "application/xml")
			_, _ = fmt.Fprint(w, listObjectsV2Response("prefix/journals/workbooks/bad.json"))
		case strings.HasSuffix(r.URL.Path, "/bucket/prefix/journals/workbooks/bad.json"):
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprint(w, "{")
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	backend := &S3Backend{
		Client: awss3.New(awss3.Options{
			Region:       "us-east-1",
			Credentials:  aws.AnonymousCredentials{},
			BaseEndpoint: aws.String(server.URL),
			UsePathStyle: true,
			HTTPClient:   server.Client(),
		}),
		Bucket: "bucket",
		Prefix: "prefix",
	}

	entries, err := backend.ListJournal("workbooks", "")
	require.Error(t, err)
	assert.Nil(t, entries)
	assert.ErrorContains(t, err, "list s3 journal workbooks completed with 1 failure")
	assert.ErrorContains(t, err, "decode journal object prefix/journals/workbooks/bad.json")
}

func TestJournalEntryMatchesQuery(t *testing.T) {
	entry := mustDecodeJournalObject(t, `{
		"id": "entry-1",
		"filename": "Plan.JSON",
		"full_path": "/tmp/projects/plan.json"
	}`)

	assert.True(t, journalEntryMatchesQuery(entry, ""))
	assert.True(t, journalEntryMatchesQuery(entry, "plan"))
	assert.True(t, journalEntryMatchesQuery(entry, "/tmp/projects"))
	assert.False(t, journalEntryMatchesQuery(entry, "notes"))
}

func mustDecodeJournalObject(t *testing.T, raw string) *model.JournalEntry {
	t.Helper()

	entry, err := decodeJournalObject("test.json", io.NopCloser(strings.NewReader(raw)))
	require.NoError(t, err)
	return entry
}

func listObjectsV2Response(key string) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<ListBucketResult xmlns="http://s3.amazonaws.com/doc/2006-03-01/">
  <Name>bucket</Name>
  <Prefix>prefix/journals/workbooks/</Prefix>
  <KeyCount>1</KeyCount>
  <MaxKeys>1000</MaxKeys>
  <IsTruncated>false</IsTruncated>
  <Contents>
    <Key>%s</Key>
    <LastModified>2026-05-30T00:00:00.000Z</LastModified>
    <ETag>"etag"</ETag>
    <Size>1</Size>
    <StorageClass>STANDARD</StorageClass>
  </Contents>
</ListBucketResult>`, key)
}
