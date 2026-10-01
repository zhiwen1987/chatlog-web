// Package store — 媒体对象存储实现（R42.7 upload bytes 级）。
//
// MinioObjectStore 是 handler.MediaObjectStore 的唯一真实实现：
//   - PutObject 流式把 reader 写入 minio 对象（分片，不整文件入内存；AGENTS A07 流式分片）
//   - 服务端同时流式计算 sha256（内容寻址幂等键）
//   - objectRef 形如 "s3://<bucket>/<tenant>/<object_name>"（与 manifest.bytes_ref 同语义）
//
// 仅做写路径；读/删/对账由外部对象生命周期流程负责，不在本工作项范围。
package store

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinioObjectStore 基于 minio-go 的对象存储写实现。
type MinioObjectStore struct {
	client *minio.Client
	bucket string
}

// NewMinioObjectStore 创建 minio 客户端；endpoint 为空时返回 (nil, err)。
func NewMinioObjectStore(endpoint, accessKey, secretKey, bucket string, useSSL bool) (*MinioObjectStore, error) {
	if endpoint == "" {
		return nil, fmt.Errorf("store: minio endpoint empty")
	}
	if bucket == "" {
		bucket = "chatlog-media"
	}
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("store: new minio client: %w", err)
	}
	return &MinioObjectStore{client: client, bucket: bucket}, nil
}

// PutObject 流式写对象并计算 sha256。返回 objectRef（s3://bucket/tenant/mediaType/name）。
func (m *MinioObjectStore) PutObject(ctx context.Context, tenantID, mediaType string, src io.Reader) (objectRef, sha256Hex string, sizeBytes int64, err error) {
	if tenantID == "" || mediaType == "" {
		return "", "", 0, fmt.Errorf("store: tenant_id/media_type required")
	}
	// bucket 幂等确保存在（已存在不报错）。
	if err := m.client.MakeBucket(ctx, m.bucket, minio.MakeBucketOptions{}); err != nil {
		exists, existsErr := m.client.BucketExists(ctx, m.bucket)
		if existsErr != nil || !exists {
			return "", "", 0, fmt.Errorf("store: ensure bucket: %w", err)
		}
	}
	key := m.objectKey(tenantID, mediaType)
	h := sha256.New()
	tee := io.TeeReader(src, h)
	n, err := m.client.PutObject(ctx, m.bucket, key, tee, -1, minio.PutObjectOptions{
		ContentType: contentTypeFor(mediaType),
	})
	if err != nil {
		return "", "", 0, fmt.Errorf("store: put object: %w", err)
	}
	sum := h.Sum(nil)
	return "s3://" + m.bucket + "/" + key, hex.EncodeToString(sum), n.Size, nil
}

// objectKey 生成对象 key：tenantID/mediaType/<时间戳>-<随机hex>。
// 内容寻址由调用方按 sha256 去重；此处 key 只需唯一、可读、租户隔离。
func (m *MinioObjectStore) objectKey(tenantID, mediaType string) string {
	return strings.Join([]string{tenantID, mediaType, fmt.Sprintf("%d-%s", time.Now().UnixNano(), randomHex(8))}, "/")
}

// randomHex 生成 n 字节随机 hex（crypto/rand；失败退化为时间戳，不崩溃）。
func randomHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

// contentTypeFor 按 media_type 映射对象 Content-Type（与协议 media-type 枚举一致）。
func contentTypeFor(mediaType string) string {
	switch mediaType {
	case "image":
		return "image/*"
	case "video":
		return "video/*"
	case "audio":
		return "audio/*"
	default:
		return "application/octet-stream"
	}
}
