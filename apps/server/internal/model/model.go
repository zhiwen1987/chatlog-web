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