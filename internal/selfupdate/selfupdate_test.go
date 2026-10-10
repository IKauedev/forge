package selfupdate_test

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ikauedev/forge/internal/selfupdate"
)

func TestCompare(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"1.2.3", "1.2.3", 0},
		{"v1.2.3", "1.2.4", -1},
		{"1.10.0", "1.9.9", 1},
		{"2.0.0", "1.99.99", 1},
		{"1.0.0-rc1", "1.0.0", 0},
		{"dev", "0.0.1", -1},
		{"1.0.0", "dev", 1},
		{"dev", "dev", 0},
	}
	for _, c := range cases {
		if got := selfupdate.Compare(c.a, c.b); got != c.want {
			t.Errorf("Compare(%q,%q) = %d, esperado %d", c.a, c.b, got, c.want)
		}
	}
}

func TestAssetName(t *testing.T) {
	if got := selfupdate.AssetName("v1.2.3", "linux", "arm64"); got != "forge_1.2.3_linux_arm64.tar.gz" {
		t.Error(got)
	}
	if got := selfupdate.AssetName("1.2.3", "windows", "amd64"); got != "forge_1.2.3_windows_amd64.zip" {
		t.Error(got)
	}
}

func tarGz(t *testing.T, name, content string) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for fname, body := range map[string]string{"README.md": "leia-me", name: content} {
		if err := tw.WriteHeader(&tar.Header{Name: fname, Mode: 0o755, Size: int64(len(body)), Typeflag: tar.TypeReg}); err != nil {
			t.Fatal(err)
		}
		_, _ = tw.Write([]byte(body))
	}
	_ = tw.Close()
	_ = gz.Close()
	return buf.Bytes()
}

func zipOf(t *testing.T, name, content string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, _ := zw.Create(name)
	_, _ = w.Write([]byte(content))
	_ = zw.Close()
	return buf.Bytes()
}

type fakeGitHub struct {
	srv       *httptest.Server
	tag       string
	files     map[string][]byte // nome do asset -> conteúdo
	checksums string            // sobrescreve o checksums.txt gerado
	token     string            // exigido, se não vazio
	hits      atomic.Int32
}

func newFake(t *testing.T, tag string, files map[string][]byte) *fakeGitHub {
	f := &fakeGitHub{tag: tag, files: files}
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/ikauedev/forge/releases/latest", func(w http.ResponseWriter, r *http.Request) {
		f.hits.Add(1)
		if f.token != "" && r.Header.Get("Authorization") != "Bearer "+f.token {
			http.NotFound(w, r) // o GitHub responde 404 a repositório privado sem credencial
			return
		}
		var assets []string
		for name := range f.files {
			assets = append(assets, fmt.Sprintf(`{"name":%q,"url":%q}`, name, f.srv.URL+"/assets/"+name))
		}
		assets = append(assets, fmt.Sprintf(`{"name":"checksums.txt","url":%q}`, f.srv.URL+"/assets/checksums.txt"))
		_, _ = fmt.Fprintf(w, `{"tag_name":%q,"assets":[%s]}`, f.tag, strings.Join(assets, ","))
	})
	mux.HandleFunc("/assets/", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Accept") != "application/octet-stream" {
			http.Error(w, "Accept errado", http.StatusBadRequest)
			return
		}
		name := strings.TrimPrefix(r.URL.Path, "/assets/")
		if name == "checksums.txt" {
			if f.checksums != "" {
				_, _ = w.Write([]byte(f.checksums))
				return
			}
			var b strings.Builder
			for n, data := range f.files {
				sum := sha256.Sum256(data)
				fmt.Fprintf(&b, "%s  %s\n", hex.EncodeToString(sum[:]), n)
			}
			_, _ = w.Write([]byte(b.String()))
			return
		}
		data, ok := f.files[name]
		if !ok {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write(data)
	})
	f.srv = httptest.NewServer(mux)
	t.Cleanup(f.srv.Close)
	return f
}

func (f *fakeGitHub) client() selfupdate.Client {
	return selfupdate.Client{API: f.srv.URL, Repo: "ikauedev/forge", Token: f.token}
}

func writeExe(t *testing.T, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "forge")
	if err := os.WriteFile(p, []byte(content), 0o755); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestUpdateReplacesTheBinaryAfterVerifyingTheChecksum(t *testing.T) {
	name := selfupdate.AssetName("1.1.0", "linux", "amd64")
	f := newFake(t, "v1.1.0", map[string][]byte{name: tarGz(t, "forge", "binario-novo")})
	exe := writeExe(t, "binario-velho")

	res, err := f.client().Update(context.Background(), "1.0.0", exe, "linux", "amd64", false)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Newer || !res.Updated || res.From != "1.0.0" || res.To != "1.1.0" {
		t.Errorf("resultado: %+v", res)
	}
	if b, _ := os.ReadFile(exe); string(b) != "binario-novo" {
		t.Errorf("binário = %q", b)
	}
	if _, err := os.Stat(exe + ".old"); err == nil {
		t.Error("sobrou o .old")
	}
}

func TestUpdateOnWindowsReadsTheZip(t *testing.T) {
	name := selfupdate.AssetName("2.0.0", "windows", "amd64")
	f := newFake(t, "v2.0.0", map[string][]byte{name: zipOf(t, "forge.exe", "exe-novo")})
	exe := writeExe(t, "exe-velho")

	res, err := f.client().Update(context.Background(), "1.9.0", exe, "windows", "amd64", false)
	if err != nil || !res.Updated {
		t.Fatalf("%+v %v", res, err)
	}
	if b, _ := os.ReadFile(exe); string(b) != "exe-novo" {
		t.Errorf("binário = %q", b)
	}
}

func TestUpdateRefusesATamperedDownload(t *testing.T) {
	name := selfupdate.AssetName("1.1.0", "linux", "amd64")
	f := newFake(t, "v1.1.0", map[string][]byte{name: tarGz(t, "forge", "malicioso")})
	f.checksums = strings.Repeat("0", 64) + "  " + name + "\n"
	exe := writeExe(t, "original")

	_, err := f.client().Update(context.Background(), "1.0.0", exe, "linux", "amd64", false)
	if err == nil || !strings.Contains(err.Error(), "SHA-256") {
		t.Fatalf("deveria recusar: %v", err)
	}
	if b, _ := os.ReadFile(exe); string(b) != "original" {
		t.Error("o binário não pode mudar quando o checksum falha")
	}
}

func TestUpdateRefusesAReleaseWithoutTheAsset(t *testing.T) {
	f := newFake(t, "v1.1.0", map[string][]byte{"forge_1.1.0_darwin_arm64.tar.gz": []byte("x")})
	_, err := f.client().Update(context.Background(), "1.0.0", writeExe(t, "x"), "linux", "amd64", false)
	if err == nil || !strings.Contains(err.Error(), "forge_1.1.0_linux_amd64.tar.gz") {
		t.Fatalf("erro = %v", err)
	}
}

func TestUpdateWhenAlreadyUpToDateOrInCheckMode(t *testing.T) {
	name := selfupdate.AssetName("1.1.0", "linux", "amd64")
	f := newFake(t, "v1.1.0", map[string][]byte{name: tarGz(t, "forge", "novo")})
	exe := writeExe(t, "atual")

	res, err := f.client().Update(context.Background(), "1.1.0", exe, "linux", "amd64", false)
	if err != nil || res.Newer || res.Updated {
		t.Errorf("já atualizado: %+v %v", res, err)
	}
	res, err = f.client().Update(context.Background(), "1.0.0", exe, "linux", "amd64", true)
	if err != nil || !res.Newer || res.Updated {
		t.Errorf("--check não pode trocar o binário: %+v %v", res, err)
	}
	if b, _ := os.ReadFile(exe); string(b) != "atual" {
		t.Error("o binário foi alterado")
	}
}

func TestDevBuildAlwaysSeesTheLatestAsNewer(t *testing.T) {
	name := selfupdate.AssetName("0.1.0", "linux", "amd64")
	f := newFake(t, "v0.1.0", map[string][]byte{name: tarGz(t, "forge", "x")})
	res, err := f.client().Update(context.Background(), selfupdate.DevVersion, writeExe(t, "dev"), "linux", "amd64", true)
	if err != nil || !res.Newer {
		t.Errorf("%+v %v", res, err)
	}
}

func TestPrivateRepositoryNeedsAToken(t *testing.T) {
	f := newFake(t, "v1.0.0", nil)
	f.token = "segredo"

	c := f.client()
	c.Token = ""
	if _, err := c.Latest(context.Background()); !errors.Is(err, selfupdate.ErrNotFound) {
		t.Errorf("sem token deveria ser ErrNotFound: %v", err)
	}
	c.Token = "segredo"
	if rel, err := c.Latest(context.Background()); err != nil || rel.Version() != "1.0.0" {
		t.Errorf("com token: %+v %v", rel, err)
	}
}

func TestCheckCachedAsksTheNetworkOncePerTTL(t *testing.T) {
	f := newFake(t, "v1.5.0", nil)
	cache := filepath.Join(t.TempDir(), "forge", "update.json")
	now := time.Date(2026, 10, 10, 12, 0, 0, 0, time.UTC)

	h, ok := selfupdate.CheckCached(context.Background(), f.client(), "1.0.0", cache, now, 24*time.Hour)
	if !ok || !h.Newer || h.Latest != "1.5.0" {
		t.Fatalf("%+v %v", h, ok)
	}
	h, ok = selfupdate.CheckCached(context.Background(), f.client(), "1.0.0", cache, now.Add(time.Hour), 24*time.Hour)
	if !ok || !h.Newer || f.hits.Load() != 1 {
		t.Errorf("deveria usar o cache (consultas=%d)", f.hits.Load())
	}
	selfupdate.CheckCached(context.Background(), f.client(), "1.0.0", cache, now.Add(25*time.Hour), 24*time.Hour)
	if f.hits.Load() != 2 {
		t.Errorf("o cache vencido deveria consultar de novo (consultas=%d)", f.hits.Load())
	}
}

func TestCheckCachedIsSilentOnFailureAndForDevBuilds(t *testing.T) {
	f := newFake(t, "v1.5.0", nil)
	f.token = "x" // sem token: 404
	c := f.client()
	c.Token = ""
	if _, ok := selfupdate.CheckCached(context.Background(), c, "1.0.0", filepath.Join(t.TempDir(), "u.json"), time.Now(), time.Hour); ok {
		t.Error("falha de rede/token não deve virar dica")
	}
	f2 := newFake(t, "v1.5.0", nil)
	if _, ok := selfupdate.CheckCached(context.Background(), f2.client(), selfupdate.DevVersion, filepath.Join(t.TempDir(), "u.json"), time.Now(), time.Hour); ok || f2.hits.Load() != 0 {
		t.Error("build dev não deve consultar a rede")
	}
}
