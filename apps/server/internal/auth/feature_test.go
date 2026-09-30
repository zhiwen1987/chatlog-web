package auth

import (
	"testing"
	"time"

	"github.com/zhiwen1987/chatlog-web/apps/server/internal/model"
)

func TestCheckFeatureAllowed(t *testing.T) {
	now := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	c := &model.Claims{
		FeatureGrants: []model.Grant{
			{GrantID: "g1", FeatureKey: "archive.read", State: model.GrantActive, ValidFrom: "2026-01-01T00:00:00Z", ValidUntil: "2027-01-01T00:00:00Z"},
		},
	}
	if d := CheckFeature(c, "archive.read", now); !d.Allowed {
		t.Fatal("archive.read should be allowed")
	}
}

func TestCheckFeatureDeniedNil(t *testing.T) {
	if d := CheckFeature(nil, "archive.read", time.Now()); d.Allowed {
		t.Fatal("nil claims must be denied (default deny)")
	}
}

func TestCheckFeatureDeniedMissingDep(t *testing.T) {
	now := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	c := &model.Claims{
		FeatureGrants: []model.Grant{
			{GrantID: "g1", FeatureKey: "media.upload.image", State: model.GrantActive, ValidFrom: "2026-01-01T00:00:00Z", ValidUntil: "2027-01-01T00:00:00Z"},
		},
	}
	// media.upload.image 依赖 archive.ingest（未授予）→ 拒绝并报告缺失依赖
	d := CheckFeature(c, "media.upload.image", now)
	if d.Allowed {
		t.Fatal("media.upload.image must be denied when dep missing")
	}
	if len(d.Missing) != 1 || d.Missing[0] != "archive.ingest" {
		t.Fatalf("Missing=%v, want [archive.ingest]", d.Missing)
	}
}

func TestCheckFeatureDeniedNoGrant(t *testing.T) {
	now := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	c := &model.Claims{FeatureGrants: nil}
	if d := CheckFeature(c, "archive.read", now); d.Allowed {
		t.Fatal("feature without grant must be denied")
	}
}
