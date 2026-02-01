// Package waldur provides Waldur API integration including object storage operations.
package waldur

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// ObjectStorageClient provides operations for Waldur object storage.
type ObjectStorageClient struct {
	client *Client
}

// NewObjectStorageClient creates a new object storage client.
func NewObjectStorageClient(c *Client) *ObjectStorageClient {
	return &ObjectStorageClient{client: c}
}

// ObjectMeta contains metadata about a stored object.
type ObjectMeta struct {
	// Key is the object key/path in the bucket
	Key string `json:"key"`

	// Size is the object size in bytes
	Size int64 `json:"size"`

	// ContentHash is the SHA-256 hash of the content
	ContentHash string `json:"content_hash"`

	// ContentType is the MIME type of the content
	ContentType string `json:"content_type"`

	// ETag is the entity tag for caching
	ETag string `json:"etag"`

	// CreatedAt is when the object was created
	CreatedAt time.Time `json:"created_at"`

	// LastModified is when the object was last modified
	LastModified time.Time `json:"last_modified"`

	// Metadata contains custom metadata
	Metadata map[string]string `json:"metadata,omitempty"`
}

// UploadRequest contains parameters for uploading an object.
type UploadRequest struct {
	// Bucket is the bucket name
	Bucket string

	// Key is the object key/path
	Key string

	// Body is the content reader
	Body io.Reader

	// Size is the expected content size (-1 if unknown)
	Size int64

	// ContentType is the MIME type
	ContentType string

	// ContentHash is the expected SHA-256 hash (optional, for verification)
	ContentHash string

	// Metadata contains custom metadata
	Metadata map[string]string
}

// UploadResponse contains the result of an upload operation.
type UploadResponse struct {
	// Key is the stored object key
	Key string `json:"key"`

	// Size is the stored object size
	Size int64 `json:"size"`

	// ContentHash is the computed SHA-256 hash
	ContentHash string `json:"content_hash"`

	// ETag is the entity tag
	ETag string `json:"etag"`

	// VersionID is the version identifier (if versioning enabled)
	VersionID string `json:"version_id,omitempty"`
}

// DownloadRequest contains parameters for downloading an object.
type DownloadRequest struct {
	// Bucket is the bucket name
	Bucket string

	// Key is the object key/path
	Key string

	// RangeStart is the byte offset to start from (optional)
	RangeStart int64

	// RangeEnd is the byte offset to end at (optional, -1 for end)
	RangeEnd int64

	// IfMatch only download if ETag matches
	IfMatch string

	// IfNoneMatch only download if ETag doesn't match
	IfNoneMatch string
}

// DownloadResponse contains the result of a download operation.
type DownloadResponse struct {
	// Body is the content reader (caller must close)
	Body io.ReadCloser

	// Size is the content size
	Size int64

	// ContentHash is the SHA-256 hash
	ContentHash string

	// ContentType is the MIME type
	ContentType string

	// ETag is the entity tag
	ETag string

	// LastModified is when the object was last modified
	LastModified time.Time
}

// ListRequest contains parameters for listing objects.
type ListRequest struct {
	// Bucket is the bucket name
	Bucket string

	// Prefix filters objects by prefix
	Prefix string

	// MaxKeys limits the number of results
	MaxKeys int

	// ContinuationToken for pagination
	ContinuationToken string
}

// ListResponse contains the result of a list operation.
type ListResponse struct {
	// Objects is the list of object metadata
	Objects []ObjectMeta `json:"objects"`

	// IsTruncated indicates if more results exist
	IsTruncated bool `json:"is_truncated"`

	// NextContinuationToken for pagination
	NextContinuationToken string `json:"next_continuation_token,omitempty"`
}

// QuotaInfo contains storage quota information.
type QuotaInfo struct {
	// TotalBytes is the total quota in bytes
	TotalBytes int64 `json:"total_bytes"`

	// UsedBytes is the currently used bytes
	UsedBytes int64 `json:"used_bytes"`

	// AvailableBytes is the remaining quota
	AvailableBytes int64 `json:"available_bytes"`

	// ObjectCount is the total number of objects
	ObjectCount int64 `json:"object_count"`

	// MaxObjectCount is the maximum allowed objects (-1 if unlimited)
	MaxObjectCount int64 `json:"max_object_count"`
}

// Object storage errors.
var (
	ErrBucketNotFound     = errors.New("bucket not found")
	ErrObjectNotFound     = errors.New("object not found")
	ErrQuotaExceeded      = errors.New("storage quota exceeded")
	ErrPreconditionFailed = errors.New("precondition failed")
	ErrUploadFailed       = errors.New("upload failed")
	ErrDownloadFailed     = errors.New("download failed")
	ErrHashMismatch       = errors.New("content hash mismatch")
)

// Upload stores an object in Waldur object storage.
func (o *ObjectStorageClient) Upload(ctx context.Context, req *UploadRequest) (*UploadResponse, error) {
	if req == nil {
		return nil, errors.New("upload request is nil")
	}
	if req.Bucket == "" {
		return nil, errors.New("bucket is required")
	}
	if req.Key == "" {
		return nil, errors.New("key is required")
	}
	if req.Body == nil {
		return nil, errors.New("body is required")
	}

	// Build the upload URL
	uploadURL := fmt.Sprintf("%s/object-storage/buckets/%s/objects/%s",
		o.client.config.BaseURL,
		url.PathEscape(req.Bucket),
		url.PathEscape(req.Key))

	var uploadResp *UploadResponse
	err := o.client.doWithRetry(ctx, func() error {
		// Read body and compute hash if streaming
		var bodyReader io.Reader
		var contentHash string
		var size int64

		if req.Size >= 0 {
			// Known size - use reader directly with hash wrapper
			hashReader := newHashingReader(req.Body)
			bodyReader = hashReader
			size = req.Size
			// Note: hash computed after read completes
		} else {
			// Unknown size - buffer first
			data, err := io.ReadAll(req.Body)
			if err != nil {
				return fmt.Errorf("read body: %w", err)
			}
			hash := sha256.Sum256(data)
			contentHash = hex.EncodeToString(hash[:])
			bodyReader = bytes.NewReader(data)
			size = int64(len(data))
		}

		httpReq, err := http.NewRequestWithContext(ctx, http.MethodPut, uploadURL, bodyReader)
		if err != nil {
			return fmt.Errorf("create request: %w", err)
		}

		// Set headers
		httpReq.Header.Set("Authorization", "Token "+o.client.config.Token)
		httpReq.Header.Set("User-Agent", o.client.config.UserAgent)
		if req.ContentType != "" {
			httpReq.Header.Set("Content-Type", req.ContentType)
		} else {
			httpReq.Header.Set("Content-Type", "application/octet-stream")
		}
		if size >= 0 {
			httpReq.ContentLength = size
		}
		if req.ContentHash != "" {
			httpReq.Header.Set("X-Content-SHA256", req.ContentHash)
		} else if contentHash != "" {
			httpReq.Header.Set("X-Content-SHA256", contentHash)
		}

		// Add custom metadata
		for k, v := range req.Metadata {
			httpReq.Header.Set("X-Object-Meta-"+k, v)
		}

		resp, err := o.client.httpClient.Do(httpReq)
		if err != nil {
			return fmt.Errorf("upload request: %w", err)
		}
		defer func() { _ = resp.Body.Close() }()

		// Handle response
		respBody, _ := io.ReadAll(resp.Body)

		switch resp.StatusCode {
		case http.StatusOK, http.StatusCreated:
			// Parse response
			var result struct {
				Key         string `json:"key"`
				Size        int64  `json:"size"`
				ContentHash string `json:"content_hash"`
				ETag        string `json:"etag"`
				VersionID   string `json:"version_id,omitempty"`
			}
			if err := json.Unmarshal(respBody, &result); err != nil {
				// If response isn't JSON, construct from headers
				result.Key = req.Key
				result.Size = size
				result.ETag = resp.Header.Get("ETag")
				result.ContentHash = resp.Header.Get("X-Content-SHA256")
			}

			// Verify hash if provided
			if req.ContentHash != "" && result.ContentHash != "" && req.ContentHash != result.ContentHash {
				return fmt.Errorf("%w: expected %s, got %s", ErrHashMismatch, req.ContentHash, result.ContentHash)
			}

			uploadResp = &UploadResponse{
				Key:         result.Key,
				Size:        result.Size,
				ContentHash: result.ContentHash,
				ETag:        result.ETag,
				VersionID:   result.VersionID,
			}
			return nil

		case http.StatusNotFound:
			return ErrBucketNotFound

		case http.StatusRequestEntityTooLarge, http.StatusInsufficientStorage:
			return ErrQuotaExceeded

		default:
			return fmt.Errorf("%w: status %d: %s", ErrUploadFailed, resp.StatusCode, string(respBody))
		}
	})

	return uploadResp, err
}

// UploadStream stores a large object using streaming upload.
func (o *ObjectStorageClient) UploadStream(ctx context.Context, req *UploadRequest) (*UploadResponse, error) {
	if req == nil {
		return nil, errors.New("upload request is nil")
	}
	if req.Bucket == "" {
		return nil, errors.New("bucket is required")
	}
	if req.Key == "" {
		return nil, errors.New("key is required")
	}
	if req.Body == nil {
		return nil, errors.New("body is required")
	}

	// Build streaming upload URL
	uploadURL := fmt.Sprintf("%s/object-storage/buckets/%s/objects/%s/stream",
		o.client.config.BaseURL,
		url.PathEscape(req.Bucket),
		url.PathEscape(req.Key))

	var uploadResp *UploadResponse
	err := o.client.doWithRetry(ctx, func() error {
		// Create pipe for streaming
		pr, pw := io.Pipe()

		// Compute hash while streaming
		hashWriter := sha256.New()
		teeReader := io.TeeReader(req.Body, hashWriter)

		// Copy to pipe in goroutine
		go func() {
			written, err := io.Copy(pw, teeReader)
			if err != nil {
				_ = pw.CloseWithError(err)
				return
			}
			_ = pw.Close()
			_ = written // size tracked by server
		}()

		httpReq, err := http.NewRequestWithContext(ctx, http.MethodPut, uploadURL, pr)
		if err != nil {
			return fmt.Errorf("create request: %w", err)
		}

		// Set headers for streaming
		httpReq.Header.Set("Authorization", "Token "+o.client.config.Token)
		httpReq.Header.Set("User-Agent", o.client.config.UserAgent)
		httpReq.Header.Set("Transfer-Encoding", "chunked")
		if req.ContentType != "" {
			httpReq.Header.Set("Content-Type", req.ContentType)
		} else {
			httpReq.Header.Set("Content-Type", "application/octet-stream")
		}

		// Add custom metadata
		for k, v := range req.Metadata {
			httpReq.Header.Set("X-Object-Meta-"+k, v)
		}

		resp, err := o.client.httpClient.Do(httpReq)
		if err != nil {
			return fmt.Errorf("streaming upload: %w", err)
		}
		defer func() { _ = resp.Body.Close() }()

		respBody, _ := io.ReadAll(resp.Body)

		switch resp.StatusCode {
		case http.StatusOK, http.StatusCreated:
			contentHash := hex.EncodeToString(hashWriter.Sum(nil))

			var result struct {
				Key         string `json:"key"`
				Size        int64  `json:"size"`
				ContentHash string `json:"content_hash"`
				ETag        string `json:"etag"`
				VersionID   string `json:"version_id,omitempty"`
			}
			if err := json.Unmarshal(respBody, &result); err != nil {
				result.Key = req.Key
				result.ETag = resp.Header.Get("ETag")
			}
			if result.ContentHash == "" {
				result.ContentHash = contentHash
			}

			// Verify hash if provided
			if req.ContentHash != "" && result.ContentHash != req.ContentHash {
				return fmt.Errorf("%w: expected %s, got %s", ErrHashMismatch, req.ContentHash, result.ContentHash)
			}

			uploadResp = &UploadResponse{
				Key:         result.Key,
				Size:        result.Size,
				ContentHash: result.ContentHash,
				ETag:        result.ETag,
				VersionID:   result.VersionID,
			}
			return nil

		case http.StatusNotFound:
			return ErrBucketNotFound

		case http.StatusRequestEntityTooLarge, http.StatusInsufficientStorage:
			return ErrQuotaExceeded

		default:
			return fmt.Errorf("%w: status %d: %s", ErrUploadFailed, resp.StatusCode, string(respBody))
		}
	})

	return uploadResp, err
}

// Download retrieves an object from Waldur object storage.
func (o *ObjectStorageClient) Download(ctx context.Context, req *DownloadRequest) (*DownloadResponse, error) {
	if req == nil {
		return nil, errors.New("download request is nil")
	}
	if req.Bucket == "" {
		return nil, errors.New("bucket is required")
	}
	if req.Key == "" {
		return nil, errors.New("key is required")
	}

	// Build download URL
	downloadURL := fmt.Sprintf("%s/object-storage/buckets/%s/objects/%s",
		o.client.config.BaseURL,
		url.PathEscape(req.Bucket),
		url.PathEscape(req.Key))

	// Rate limit
	if err := o.client.rateLimiter.Wait(ctx); err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Authorization", "Token "+o.client.config.Token)
	httpReq.Header.Set("User-Agent", o.client.config.UserAgent)

	// Range request
	if req.RangeStart > 0 || req.RangeEnd > 0 {
		rangeHeader := fmt.Sprintf("bytes=%d-", req.RangeStart)
		if req.RangeEnd > 0 {
			rangeHeader = fmt.Sprintf("bytes=%d-%d", req.RangeStart, req.RangeEnd)
		}
		httpReq.Header.Set("Range", rangeHeader)
	}

	// Conditional headers
	if req.IfMatch != "" {
		httpReq.Header.Set("If-Match", req.IfMatch)
	}
	if req.IfNoneMatch != "" {
		httpReq.Header.Set("If-None-Match", req.IfNoneMatch)
	}

	resp, err := o.client.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("download request: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusOK, http.StatusPartialContent:
		// Success - return streaming body
		var lastModified time.Time
		if lm := resp.Header.Get("Last-Modified"); lm != "" {
			lastModified, _ = time.Parse(time.RFC1123, lm)
		}

		return &DownloadResponse{
			Body:         resp.Body,
			Size:         resp.ContentLength,
			ContentHash:  resp.Header.Get("X-Content-SHA256"),
			ContentType:  resp.Header.Get("Content-Type"),
			ETag:         resp.Header.Get("ETag"),
			LastModified: lastModified,
		}, nil

	case http.StatusNotFound:
		_ = resp.Body.Close()
		return nil, ErrObjectNotFound

	case http.StatusPreconditionFailed:
		_ = resp.Body.Close()
		return nil, ErrPreconditionFailed

	case http.StatusNotModified:
		_ = resp.Body.Close()
		return nil, ErrPreconditionFailed

	default:
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		return nil, fmt.Errorf("%w: status %d: %s", ErrDownloadFailed, resp.StatusCode, string(body))
	}
}

// Head retrieves object metadata without downloading content.
func (o *ObjectStorageClient) Head(ctx context.Context, bucket, key string) (*ObjectMeta, error) {
	if bucket == "" {
		return nil, errors.New("bucket is required")
	}
	if key == "" {
		return nil, errors.New("key is required")
	}

	headURL := fmt.Sprintf("%s/object-storage/buckets/%s/objects/%s",
		o.client.config.BaseURL,
		url.PathEscape(bucket),
		url.PathEscape(key))

	var meta *ObjectMeta
	err := o.client.doWithRetry(ctx, func() error {
		httpReq, err := http.NewRequestWithContext(ctx, http.MethodHead, headURL, nil)
		if err != nil {
			return fmt.Errorf("create request: %w", err)
		}

		httpReq.Header.Set("Authorization", "Token "+o.client.config.Token)
		httpReq.Header.Set("User-Agent", o.client.config.UserAgent)

		resp, err := o.client.httpClient.Do(httpReq)
		if err != nil {
			return fmt.Errorf("head request: %w", err)
		}
		defer func() { _ = resp.Body.Close() }()

		switch resp.StatusCode {
		case http.StatusOK:
			var lastModified time.Time
			if lm := resp.Header.Get("Last-Modified"); lm != "" {
				lastModified, _ = time.Parse(time.RFC1123, lm)
			}

			metadata := make(map[string]string)
			for k, v := range resp.Header {
				if strings.HasPrefix(k, "X-Object-Meta-") && len(v) > 0 {
					metadata[strings.TrimPrefix(k, "X-Object-Meta-")] = v[0]
				}
			}

			meta = &ObjectMeta{
				Key:          key,
				Size:         resp.ContentLength,
				ContentHash:  resp.Header.Get("X-Content-SHA256"),
				ContentType:  resp.Header.Get("Content-Type"),
				ETag:         resp.Header.Get("ETag"),
				LastModified: lastModified,
				Metadata:     metadata,
			}
			return nil

		case http.StatusNotFound:
			return ErrObjectNotFound

		default:
			return fmt.Errorf("head failed: status %d", resp.StatusCode)
		}
	})

	return meta, err
}

// Exists checks if an object exists.
func (o *ObjectStorageClient) Exists(ctx context.Context, bucket, key string) (bool, error) {
	_, err := o.Head(ctx, bucket, key)
	if err != nil {
		if errors.Is(err, ErrObjectNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// Delete removes an object from storage.
func (o *ObjectStorageClient) Delete(ctx context.Context, bucket, key string) error {
	if bucket == "" {
		return errors.New("bucket is required")
	}
	if key == "" {
		return errors.New("key is required")
	}

	deleteURL := fmt.Sprintf("%s/object-storage/buckets/%s/objects/%s",
		o.client.config.BaseURL,
		url.PathEscape(bucket),
		url.PathEscape(key))

	return o.client.doWithRetry(ctx, func() error {
		httpReq, err := http.NewRequestWithContext(ctx, http.MethodDelete, deleteURL, nil)
		if err != nil {
			return fmt.Errorf("create request: %w", err)
		}

		httpReq.Header.Set("Authorization", "Token "+o.client.config.Token)
		httpReq.Header.Set("User-Agent", o.client.config.UserAgent)

		resp, err := o.client.httpClient.Do(httpReq)
		if err != nil {
			return fmt.Errorf("delete request: %w", err)
		}
		defer func() { _ = resp.Body.Close() }()

		switch resp.StatusCode {
		case http.StatusOK, http.StatusNoContent:
			return nil

		case http.StatusNotFound:
			// Object already doesn't exist - not an error
			return nil

		default:
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("delete failed: status %d: %s", resp.StatusCode, string(body))
		}
	})
}

// List retrieves a list of objects matching the prefix.
func (o *ObjectStorageClient) List(ctx context.Context, req *ListRequest) (*ListResponse, error) {
	if req == nil {
		return nil, errors.New("list request is nil")
	}
	if req.Bucket == "" {
		return nil, errors.New("bucket is required")
	}

	// Build list URL with query parameters
	listURL := fmt.Sprintf("%s/object-storage/buckets/%s/objects",
		o.client.config.BaseURL,
		url.PathEscape(req.Bucket))

	params := url.Values{}
	if req.Prefix != "" {
		params.Set("prefix", req.Prefix)
	}
	if req.MaxKeys > 0 {
		params.Set("max_keys", strconv.Itoa(req.MaxKeys))
	}
	if req.ContinuationToken != "" {
		params.Set("continuation_token", req.ContinuationToken)
	}
	if len(params) > 0 {
		listURL += "?" + params.Encode()
	}

	var listResp *ListResponse
	err := o.client.doWithRetry(ctx, func() error {
		httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, listURL, nil)
		if err != nil {
			return fmt.Errorf("create request: %w", err)
		}

		httpReq.Header.Set("Authorization", "Token "+o.client.config.Token)
		httpReq.Header.Set("User-Agent", o.client.config.UserAgent)
		httpReq.Header.Set("Accept", "application/json")

		resp, err := o.client.httpClient.Do(httpReq)
		if err != nil {
			return fmt.Errorf("list request: %w", err)
		}
		defer func() { _ = resp.Body.Close() }()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("read response: %w", err)
		}

		switch resp.StatusCode {
		case http.StatusOK:
			if err := json.Unmarshal(body, &listResp); err != nil {
				return fmt.Errorf("parse response: %w", err)
			}
			return nil

		case http.StatusNotFound:
			return ErrBucketNotFound

		default:
			return fmt.Errorf("list failed: status %d: %s", resp.StatusCode, string(body))
		}
	})

	return listResp, err
}

// GetQuota retrieves storage quota information for the bucket.
func (o *ObjectStorageClient) GetQuota(ctx context.Context, bucket string) (*QuotaInfo, error) {
	if bucket == "" {
		return nil, errors.New("bucket is required")
	}

	quotaURL := fmt.Sprintf("%s/object-storage/buckets/%s/quota",
		o.client.config.BaseURL,
		url.PathEscape(bucket))

	var quota *QuotaInfo
	err := o.client.doWithRetry(ctx, func() error {
		httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, quotaURL, nil)
		if err != nil {
			return fmt.Errorf("create request: %w", err)
		}

		httpReq.Header.Set("Authorization", "Token "+o.client.config.Token)
		httpReq.Header.Set("User-Agent", o.client.config.UserAgent)
		httpReq.Header.Set("Accept", "application/json")

		resp, err := o.client.httpClient.Do(httpReq)
		if err != nil {
			return fmt.Errorf("quota request: %w", err)
		}
		defer func() { _ = resp.Body.Close() }()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("read response: %w", err)
		}

		switch resp.StatusCode {
		case http.StatusOK:
			if err := json.Unmarshal(body, &quota); err != nil {
				return fmt.Errorf("parse quota: %w", err)
			}
			return nil

		case http.StatusNotFound:
			return ErrBucketNotFound

		default:
			return fmt.Errorf("quota failed: status %d: %s", resp.StatusCode, string(body))
		}
	})

	return quota, err
}

// CheckQuota verifies if the required size can be stored within quota.
func (o *ObjectStorageClient) CheckQuota(ctx context.Context, bucket string, requiredBytes int64) error {
	quota, err := o.GetQuota(ctx, bucket)
	if err != nil {
		return err
	}

	if quota.AvailableBytes < requiredBytes {
		return fmt.Errorf("%w: need %d bytes, have %d available", ErrQuotaExceeded, requiredBytes, quota.AvailableBytes)
	}

	return nil
}

// Ping verifies the object storage service is accessible.
func (o *ObjectStorageClient) Ping(ctx context.Context) error {
	pingURL := fmt.Sprintf("%s/object-storage/health", o.client.config.BaseURL)

	return o.client.doWithRetry(ctx, func() error {
		httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, pingURL, nil)
		if err != nil {
			return fmt.Errorf("create request: %w", err)
		}

		httpReq.Header.Set("Authorization", "Token "+o.client.config.Token)
		httpReq.Header.Set("User-Agent", o.client.config.UserAgent)

		resp, err := o.client.httpClient.Do(httpReq)
		if err != nil {
			return fmt.Errorf("ping request: %w", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return nil
		}

		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("health check failed: status %d: %s", resp.StatusCode, string(body))
	})
}

// hashingReader wraps a reader and computes SHA-256 hash while reading.
type hashingReader struct {
	reader io.Reader
	hasher io.Writer
	hash   []byte
}

func newHashingReader(r io.Reader) *hashingReader {
	h := sha256.New()
	return &hashingReader{
		reader: io.TeeReader(r, h),
		hasher: h,
	}
}

func (h *hashingReader) Read(p []byte) (n int, err error) {
	return h.reader.Read(p)
}

func (h *hashingReader) Hash() string {
	if hasher, ok := h.hasher.(interface{ Sum([]byte) []byte }); ok {
		return hex.EncodeToString(hasher.Sum(nil))
	}
	return ""
}

