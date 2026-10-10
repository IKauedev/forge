package project

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestWriteReadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	want := Config{
		Forge:    ForgeInfo{Version: "0.1.0"},
		Template: "spring-api",
		Project:  ProjectInfo{Name: "Payments API", Package: "com.empresa"},
		Options:  map[string]string{"http": "both", "auth": "jwt", "cache": "none"},
	}
	if err := Write(dir, want); err != nil {
		t.Fatal(err)
	}
	got, err := Read(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
	raw, _ := os.ReadFile(filepath.Join(dir, File))
	if !strings.HasPrefix(string(raw), "# Gerado pelo forge") {
		t.Errorf("faltou cabeçalho: %s", raw)
	}
}

func TestReadMissingAndInvalid(t *testing.T) {
	dir := t.TempDir()
	if _, err := Read(dir); err == nil {
		t.Error("esperava erro para arquivo ausente")
	}
	if err := os.WriteFile(filepath.Join(dir, File), []byte("forge: [quebrado"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Read(dir); err == nil {
		t.Error("esperava erro para yaml inválido")
	}
}
