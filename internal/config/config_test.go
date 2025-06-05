package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAppConfig_FileNotFound(t *testing.T) {
	dir := t.TempDir()
	prev, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(prev); err != nil {
			t.Fatalf("restore wd: %v", err)
		}
	})
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}

	cfg, err := GetAppConfig()
	assert.Error(t, err)
	assert.Nil(t, cfg)
}

func TestGetAppConfig_SuccessWithEnvOverride(t *testing.T) {
	dir := t.TempDir()
	prev, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(prev); err != nil {
			t.Fatalf("restore wd: %v", err)
		}
	})
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}

	data := []byte("api:\n  bind: \":8080\"\ndb:\n  conn_string: \"default\"\n  max_open_cons: 10\n  conn_max_lifetime: 5s\n  migration_dir_path: \"./migrations\"\n  migration_table: \"migrations\"\n")
	if err := os.WriteFile(filepath.Join(dir, "config.yml"), data, 0o644); err != nil {
		t.Fatalf("write config failed: %v", err)
	}

	cfg, err := GetAppConfig()
	if err != nil {
		t.Fatalf("GetAppConfig returned error: %v", err)
	}
	assert.Equal(t, ":8080", cfg.API.Bind)
	assert.Equal(t, "default", cfg.Db.ConnString)
	assert.Equal(t, 10, cfg.Db.MaxOpenCons)
	assert.Equal(t, "migrations", cfg.Db.MigrationTable)
}

func TestGetAppConfig_UnmarshalError(t *testing.T) {
	dir := t.TempDir()
	prev, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(prev); err != nil {
			t.Fatalf("restore wd: %v", err)
		}
	})
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}

	data := []byte("api:\n  bind: \":8080\"\ndb:\n  max_open_cons: \"notint\"\n")
	if err := os.WriteFile(filepath.Join(dir, "config.yml"), data, 0o644); err != nil {
		t.Fatalf("write config failed: %v", err)
	}

	cfg, err := GetAppConfig()
	assert.Error(t, err)
	assert.Nil(t, cfg)
}
