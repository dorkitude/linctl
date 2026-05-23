package auth

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func withTempHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	return home
}

func withMockPass(t *testing.T, fn func(stdin io.Reader, args ...string) ([]byte, error)) {
	t.Helper()
	orig := runPassCommand
	runPassCommand = fn
	t.Cleanup(func() {
		runPassCommand = orig
	})
}

func writeStdin(t *testing.T, value string) func() {
	t.Helper()
	orig := os.Stdin
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("create stdin pipe: %v", err)
	}
	if _, err := writer.WriteString(value); err != nil {
		t.Fatalf("write stdin pipe: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close stdin writer: %v", err)
	}
	os.Stdin = reader
	return func() {
		os.Stdin = orig
		_ = reader.Close()
	}
}

func mockViewer(t *testing.T) func() {
	t.Helper()
	orig := http.DefaultTransport
	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if got := req.Header.Get("Authorization"); got != "new-key" {
			t.Fatalf("expected auth header new-key, got %q", got)
		}

		var gqlReq struct {
			Query string `json:"query"`
		}
		if err := json.NewDecoder(req.Body).Decode(&gqlReq); err != nil {
			t.Fatalf("decode GraphQL request: %v", err)
		}
		if !strings.Contains(gqlReq.Query, "viewer") {
			t.Fatalf("expected viewer query, got %q", gqlReq.Query)
		}

		body := `{"data":{"viewer":{"id":"u1","name":"Test User","email":"test@example.com","avatarUrl":"","isMe":true,"active":true,"admin":false}}}`
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(body)),
		}, nil
	})
	return func() {
		http.DefaultTransport = orig
	}
}

func TestGetAuthHeaderEnvOverridesPassAndConfig(t *testing.T) {
	withTempHome(t)
	t.Setenv("LINCTL_API_KEY", "env-key")
	t.Setenv("LINCTL_PASS_NAME", "linear-api-key")
	if err := saveAuth(AuthConfig{APIKey: "config-key"}); err != nil {
		t.Fatalf("saveAuth: %v", err)
	}

	withMockPass(t, func(stdin io.Reader, args ...string) ([]byte, error) {
		t.Fatalf("pass should not be called when LINCTL_API_KEY is set")
		return nil, nil
	})

	got, err := GetAuthHeader()
	if err != nil {
		t.Fatalf("GetAuthHeader returned error: %v", err)
	}
	if got != "env-key" {
		t.Fatalf("expected env-key, got %q", got)
	}
}

func TestGetAuthHeaderReadsFirstPassLineWithOptionTerminator(t *testing.T) {
	withTempHome(t)
	t.Setenv("LINCTL_PASS_NAME", "-linear-api-key")

	withMockPass(t, func(stdin io.Reader, args ...string) ([]byte, error) {
		want := []string{"show", "--", "-linear-api-key"}
		if !reflect.DeepEqual(args, want) {
			t.Fatalf("expected args %v, got %v", want, args)
		}
		return []byte("lin_api_x\nnotes should not be part of the header\n"), nil
	})

	got, err := GetAuthHeader()
	if err != nil {
		t.Fatalf("GetAuthHeader returned error: %v", err)
	}
	if got != "lin_api_x" {
		t.Fatalf("expected first pass line, got %q", got)
	}
}

func TestLoginWithPassRemovesLegacyConfig(t *testing.T) {
	home := withTempHome(t)
	t.Setenv("LINCTL_PASS_NAME", "linear-api-key")
	if err := saveAuth(AuthConfig{APIKey: "old-config-key"}); err != nil {
		t.Fatalf("saveAuth: %v", err)
	}
	defer mockViewer(t)()
	defer writeStdin(t, "new-key\n")()

	passInsertCalled := false
	withMockPass(t, func(stdin io.Reader, args ...string) ([]byte, error) {
		want := []string{"insert", "-m", "-f", "--", "linear-api-key"}
		if !reflect.DeepEqual(args, want) {
			t.Fatalf("expected args %v, got %v", want, args)
		}
		data, err := io.ReadAll(stdin)
		if err != nil {
			t.Fatalf("read pass stdin: %v", err)
		}
		if string(data) != "new-key\n" {
			t.Fatalf("expected pass stdin to contain API key only, got %q", string(data))
		}
		passInsertCalled = true
		return []byte{}, nil
	})

	if err := Login(true, false); err != nil {
		t.Fatalf("Login returned error: %v", err)
	}
	if !passInsertCalled {
		t.Fatalf("expected pass insert to be called")
	}
	if _, err := os.Stat(filepath.Join(home, ".linctl-auth.json")); !os.IsNotExist(err) {
		t.Fatalf("expected legacy auth config to be removed, stat err=%v", err)
	}
}

func TestLogoutRemovesConfigEvenWhenPassRemoveFails(t *testing.T) {
	home := withTempHome(t)
	t.Setenv("LINCTL_PASS_NAME", "linear-api-key")
	if err := saveAuth(AuthConfig{APIKey: "old-config-key"}); err != nil {
		t.Fatalf("saveAuth: %v", err)
	}

	withMockPass(t, func(stdin io.Reader, args ...string) ([]byte, error) {
		want := []string{"rm", "-f", "--", "linear-api-key"}
		if !reflect.DeepEqual(args, want) {
			t.Fatalf("expected args %v, got %v", want, args)
		}
		return []byte("pass entry missing"), errors.New("pass failed")
	})

	if err := Logout(); err == nil {
		t.Fatalf("expected Logout to report pass rm failure")
	}
	if _, err := os.Stat(filepath.Join(home, ".linctl-auth.json")); !os.IsNotExist(err) {
		t.Fatalf("expected legacy auth config to be removed despite pass failure, stat err=%v", err)
	}
}
