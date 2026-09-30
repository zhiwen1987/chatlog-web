package model

import (
	"testing"
	"time"
)

func testClaims() *Claims {
	// 合成 claims：archive.read + archive.ingest active，media.upload.image active 依赖 ingest
	return &Claims{
		Typ: "wcm.license", Iss: "issuer-test", Aud: "product", Licensee: "l1",
		Deployment: "d1", FeatureCatalogVer: 1, LicenseRevision: 1,
		FeatureGrants: []Grant{
			{GrantID: "g1", FeatureKey: "archive.read", State: GrantActive, ValidFrom: "2026-01-01T00:00:00Z", ValidUntil: "2027-01-01T00:00:00Z"},
			{GrantID: "g2", FeatureKey: "archive.ingest", State: GrantActive, ValidFrom: "2026-01-01T00:00:00Z", ValidUntil: "2027-01-01T00:00:00Z"},
			{GrantID: "g3", FeatureKey: "media.upload.image", State: GrantActive, ValidFrom: "2026-01-01T00:00:00Z", ValidUntil: "2027-01-01T00:00:00Z"},
		},
	}
}

func TestHasFeatureActive(t *testing.T) {
	now := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	c := testClaims()
	if !c.HasFeature("archive.read", now) {
		t.Fatal("archive.read should be granted")
	}
	if !c.HasFeature("media.upload.image", now) {
		t.Fatal("media.upload.image should be granted (dep archive.ingest active)")
	}
}

func TestHasFeatureDependencyMissing(t *testing.T) {
	now := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	c := testClaims()
	// 移除 archive.ingest 的 grant → media.upload.image 依赖不满足
	var grants []Grant
	for _, g := range c.FeatureGrants {
		if g.FeatureKey != "archive.ingest" {
			grants = append(grants, g)
		}
	}
	c.FeatureGrants = grants
	if c.HasFeature("media.upload.image", now) {
		t.Fatal("media.upload.image must fail when dep archive.ingest missing")
	}
	// MissingDeps 报告缺失依赖
	md := c.MissingDeps("media.upload.image", now)
	if len(md) != 1 || md[0] != "archive.ingest" {
		t.Fatalf("MissingDeps = %v, want [archive.ingest]", md)
	}
}

func TestHasFeatureStates(t *testing.T) {
	now := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	c := testClaims()
	for _, st := range []GrantState{GrantPaused, GrantRevoked, GrantExpired} {
		c.FeatureGrants[0].State = st
		if c.HasFeature("archive.read", now) {
			t.Fatalf("archive.read must be denied when state=%s", st)
		}
	}
}

func TestGrantWindowBoundary(t *testing.T) {
	c := testClaims()
	// 左闭：valid_from 当日 00:00 应有效
	atFrom := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	if !c.grantFor("archive.read").grantActive(atFrom) {
		t.Fatal("left-closed: valid_from boundary should be active")
	}
	// 右开：valid_until 当日 00:00 应失效
	atUntil := time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)
	if c.grantFor("archive.read").grantActive(atUntil) {
		t.Fatal("right-open: valid_until boundary should be inactive")
	}
}

func TestHasFeatureDependencyCycle(t *testing.T) {
	now := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	// 模拟损坏目录：把 FeatureDeps 临时改为 a→b→a（test 后恢复）
	origA, origB := FeatureDeps["media.upload.image"], FeatureDeps["media.transcode"]
	FeatureDeps["media.upload.image"] = []string{"media.transcode"}
	FeatureDeps["media.transcode"] = []string{"media.upload.image"}
	defer func() { FeatureDeps["media.upload.image"] = origA; FeatureDeps["media.transcode"] = origB }()
	c := testClaims()
	if c.HasFeature("media.upload.image", now) {
		t.Fatal("cycle must be treated as unsatisfied, not infinite recursion")
	}
}

func TestNilOrEmptyClaims(t *testing.T) {
	now := time.Now()
	var c *Claims
	if c.HasFeature("archive.read", now) {
		t.Fatal("nil claims must deny")
	}
	empty := &Claims{}
	if empty.HasFeature("archive.read", now) {
		t.Fatal("empty grants must deny")
	}
}
