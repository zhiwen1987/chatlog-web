package model

import "time"

// 角色（规格 §63）
const (
	RoleOwner       = "owner"
	RoleAdmin       = "admin"
	RoleSupervisor  = "supervisor"
	RoleEmployee    = "employee"
	RoleAnalyst     = "analyst"
	RoleAuditor     = "auditor"
	RoleDeviceAgent = "device_agent"
)

// 角色权限权重（数字越小权限越大，用于比较）。
var RoleWeight = map[string]int{
	RoleOwner:       0,
	RoleAdmin:       1,
	RoleSupervisor:  2,
	RoleEmployee:    3,
	RoleAnalyst:     4,
	RoleAuditor:     5,
	RoleDeviceAgent: 6,
}

// Tenant 企业租户。
type Tenant struct {
	ID        string
	Name      string
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// User 后台用户。
type User struct {
	ID           string
	Email        string
	Phone        string
	Name         string
	PasswordHash string
	Status       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// TenantMember 租户成员关系。
type TenantMember struct {
	ID        string
	TenantID  string
	UserID    string
	Role      string
	TeamID    string
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Device 设备。
type Device struct {
	ID                string
	TenantID          string
	UserID            string
	DeviceName        string
	DeviceFingerprint string
	Platform          string
	PlatformVersion   string
	Architecture      string
	ClientVersion     string
	PublicKey         string
	Status            string
	FirstSeenAt       time.Time
	LastSeenAt        time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// SourceAccount 数据来源账号。
type SourceAccount struct {
	ID                 string
	TenantID           string
	SourceType         string
	ExternalAccountID  string
	DisplayName        string
	PrimaryDeviceID    string
	Status             string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// SourceShard 数据分片。
type SourceShard struct {
	ID                string
	TenantID          string
	SourceAccountID   string
	LogicalName       string
	SchemaFingerprint string
	SourceIdentity    string
	Status            string
	LastSeenAt        time.Time
	CreatedAt         time.Time
}
// MediaOrigin 媒体来源（media-manifest 协议 origin 对象）。
type MediaOrigin struct {
	Kind              string `json:"kind"`
	Source            string `json:"source"`
	SourceMessageRef  string `json:"source_message_ref"`
}

// MediaManifest 媒体清单（R42.7，对齐 packages/protocol/media-manifest.json）。
type MediaManifest struct {
	ManifestID     string      `json:"manifest_id"`
	CatalogVersion int         `json:"catalog_version"`
	TenantID       string      `json:"tenant_id"`
	DeploymentID   string      `json:"deployment_id"`
	ObjectRef      string      `json:"object_ref"`
	MediaType      string      `json:"media_type"`
	Origin         MediaOrigin `json:"origin"`
	BytesRef       string      `json:"bytes_ref"`
	SHA256         string      `json:"sha256"`
	SizeBytes      int64       `json:"size_bytes"`
	Seq            string      `json:"seq"`
	CreatedAt      string      `json:"created_at"`
	State          string      `json:"state"`
	Replaces       string      `json:"replaces,omitempty"`
}

// MediaReceipt 媒体收据（R42.7，对齐 packages/protocol/media-receipt.json）。
type MediaReceipt struct {
	ReceiptID        string `json:"receipt_id"`
	CatalogVersion   int    `json:"catalog_version"`
	TenantID         string `json:"tenant_id"`
	DeploymentID     string `json:"deployment_id"`
	Source           string `json:"source"`
	SourceMessageRef string `json:"source_message_ref"`
	Seq              string `json:"seq"`
	CommittedAt      string `json:"committed_at"`
	BackupSet        string `json:"backup_set"`
	ObjectRef        string `json:"object_ref,omitempty"`
	MediaKind        string `json:"media_kind,omitempty"`
}

// IngestMessage 数据链接入消息（R42 ingestion 入参）。
// 对齐 messages 表字段 + 来源/会话引用；方向/类型/解码状态用枚举（与 messages CHECK 一致）。
type IngestMessage struct {
	UpstreamMessageID string `json:"upstream_message_id"`
	SourceType        string `json:"source_type"`         // chatlog_http/plaintext_wechat_sqlite/...（对齐 source_accounts.source_type）
	SourceExternalID  string `json:"source_external_id"`
	SourceDisplayName string `json:"source_display_name"`
	ConversationRef   string `json:"conversation_ref"`   // external_conversation_id
	ConversationType  string `json:"conversation_type"`  // private/group/system
	ConversationTitle string `json:"conversation_title,omitempty"`
	Direction         string `json:"direction"`          // incoming/outgoing/unknown
	MessageType       string `json:"message_type"`       // text/image/video/voice/file/emoji/system/link/quote/unknown
	SentAt            any    `json:"sent_at,omitempty"`  // RFC3339 或 null
	SentAtMs          *int64 `json:"sent_at_ms,omitempty"`
	ContentText       string `json:"content_text,omitempty"`
	ContentJSON       any    `json:"content_json,omitempty"`
	ContentHash       string `json:"content_hash,omitempty"`
	DecodeStatus      string `json:"decode_status"`      // ok/partial/failed/encrypted
}
