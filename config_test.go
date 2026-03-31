package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseConfig_shouldParseAllCliArgs(t *testing.T) {
	cfg, err := parseConfig([]string{
		"-e", "https://api.example.com",
		"-a", "myApp",
		"-u", "admin",
		"-p", "secret",
		"-d", "dir1",
		"-k",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertEqual(t, "Endpoint", "https://api.example.com", cfg.Endpoint)
	assertEqual(t, "Application", "myApp", cfg.Application)
	assertEqual(t, "User", "admin", cfg.User)
	assertEqual(t, "Password", "secret", cfg.Password)
	assertEqual(t, "Directory", "dir1", cfg.Directory)
}

func TestParseConfig_shouldReturnEmptyStringsWhenNoArgsProvided(t *testing.T) {
	cfg, err := parseConfig([]string{"--cleanup"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertEqual(t, "Endpoint", "", cfg.Endpoint)
	assertEqual(t, "Application", "", cfg.Application)
	assertEqual(t, "User", "", cfg.User)
	assertEqual(t, "Password", "", cfg.Password)
	assertEqual(t, "Directory", "", cfg.Directory)
}

func TestParseConfig_shouldParsePartialCliArgs(t *testing.T) {
	cfg, err := parseConfig([]string{"-e", "https://api.example.com", "-u", "admin", "-k"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertEqual(t, "Endpoint", "https://api.example.com", cfg.Endpoint)
	assertEqual(t, "Application", "", cfg.Application)
	assertEqual(t, "User", "admin", cfg.User)
	assertEqual(t, "Password", "", cfg.Password)
	assertEqual(t, "Directory", "", cfg.Directory)
}

func TestParseConfig_shouldLoadConfigFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.properties")
	content := strings.Join([]string{
		"endpoint=https://from-file.example.com",
		"application=fileApp",
		"user=fileUser",
		"password=filePass",
		"directory=fileDir",
	}, "\n")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	cfg, err := parseConfig([]string{"-c", path, "-k"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertEqual(t, "Endpoint", "https://from-file.example.com", cfg.Endpoint)
	assertEqual(t, "Application", "fileApp", cfg.Application)
	assertEqual(t, "User", "fileUser", cfg.User)
	assertEqual(t, "Password", "filePass", cfg.Password)
	assertEqual(t, "Directory", "fileDir", cfg.Directory)
}

func TestParseConfig_shouldAllowCliArgsToOverrideConfigFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.properties")
	content := strings.Join([]string{
		"endpoint=https://from-file.example.com",
		"user=fileUser",
		"password=filePass",
	}, "\n")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	cfg, err := parseConfig([]string{
		"-c", path,
		"-e", "https://from-cli.example.com",
		"-u", "cliUser",
		"-k",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertEqual(t, "Endpoint", "https://from-cli.example.com", cfg.Endpoint)
	assertEqual(t, "User", "cliUser", cfg.User)
	assertEqual(t, "Password", "filePass", cfg.Password)
}

func TestParseConfig_shouldHandlePartialConfigFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.properties")
	if err := os.WriteFile(path, []byte("endpoint=https://from-file.example.com\n"), 0600); err != nil {
		t.Fatal(err)
	}

	cfg, err := parseConfig([]string{"-c", path, "-k"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertEqual(t, "Endpoint", "https://from-file.example.com", cfg.Endpoint)
	assertEqual(t, "Application", "", cfg.Application)
	assertEqual(t, "User", "", cfg.User)
	assertEqual(t, "Password", "", cfg.Password)
	assertEqual(t, "Directory", "", cfg.Directory)
}

func TestParseConfig_shouldReturnErrorOnInvalidConfigFile(t *testing.T) {
	_, err := parseConfig([]string{"-c", "/nonexistent/path/config.properties", "-k"})
	if err == nil {
		t.Fatal("expected error for nonexistent config file")
	}
}

func TestValidate_shouldPassWhenAllFieldsPresent(t *testing.T) {
	cfg := AppConfig{
		Endpoint:    "https://api.example.com",
		Application: "myApp",
		User:        "admin",
		Password:    "secret",
		Directory:   "dir1",
		PageSize:    DefaultPageSize,
	}
	if err := cfg.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidate_shouldErrorWhenEndpointMissing(t *testing.T) {
	cfg := AppConfig{Application: "myApp", User: "admin", Password: "secret", Directory: "dir1", PageSize: DefaultPageSize}
	err := cfg.Validate()
	assertError(t, "endpoint is required", err)
}

func TestValidate_shouldPassWhenEndpointValid(t *testing.T) {
	cfg := AppConfig{
		Endpoint: "http://server.com/api", Application: "myApp",
		User: "admin", Password: "secret", Directory: "dir1", PageSize: DefaultPageSize,
	}
	if err := cfg.Validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidate_shouldErrorWhenEndpointInvalid(t *testing.T) {
	cfg := AppConfig{
		Endpoint: "http://invalid com", Application: "myApp",
		User: "admin", Password: "secret", Directory: "dir1", PageSize: DefaultPageSize,
	}
	err := cfg.Validate()
	assertError(t, "endpoint must be a valid URL", err)
}

func TestValidate_shouldErrorWhenApplicationMissing(t *testing.T) {
	cfg := AppConfig{
		Endpoint: "https://api.example.com", Application: "",
		User: "admin", Password: "secret", Directory: "dir1", PageSize: DefaultPageSize,
	}
	err := cfg.Validate()
	assertError(t, "application is required", err)
}

func TestValidate_shouldErrorWhenUserMissing(t *testing.T) {
	cfg := AppConfig{
		Endpoint: "https://api.example.com", Application: "myApp",
		User: "  ", Password: "secret", Directory: "dir1", PageSize: DefaultPageSize,
	}
	err := cfg.Validate()
	assertError(t, "user is required", err)
}

func TestValidate_shouldErrorWhenPasswordMissing(t *testing.T) {
	cfg := AppConfig{
		Endpoint: "https://api.example.com", Application: "myApp",
		User: "admin", Password: "", Directory: "dir1", PageSize: DefaultPageSize,
	}
	err := cfg.Validate()
	assertError(t, "password is required", err)
}

func TestValidate_shouldErrorWhenDirectoryMissing(t *testing.T) {
	cfg := AppConfig{
		Endpoint: "https://api.example.com", Application: "myApp",
		User: "admin", Password: "secret", Directory: "", PageSize: DefaultPageSize,
	}
	err := cfg.Validate()
	assertError(t, "directory is required", err)
}

func TestParseConfig_shouldAcceptLongOptionNames(t *testing.T) {
	cfg, err := parseConfig([]string{
		"--endpoint", "https://api.example.com",
		"--application", "myApp",
		"--user", "admin",
		"--password", "secret",
		"--directory", "dir1",
		"--cleanup",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertEqual(t, "Endpoint", "https://api.example.com", cfg.Endpoint)
	assertEqual(t, "Application", "myApp", cfg.Application)
	assertEqual(t, "User", "admin", cfg.User)
	assertEqual(t, "Password", "secret", cfg.Password)
	assertEqual(t, "Directory", "dir1", cfg.Directory)
}

// helpers

func assertEqual(t *testing.T, field, want, got string) {
	t.Helper()
	if want != got {
		t.Errorf("%s: expected %q, got %q", field, want, got)
	}
}

func assertError(t *testing.T, wantMsg string, err error) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error %q, got nil", wantMsg)
	}
	if err.Error() != wantMsg {
		t.Errorf("expected error %q, got %q", wantMsg, err.Error())
	}
}
