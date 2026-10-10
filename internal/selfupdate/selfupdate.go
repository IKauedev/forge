// Package selfupdate atualiza o binário do forge a partir das releases do GitHub.
package selfupdate

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// DefaultAPI é a API pública do GitHub.
const DefaultAPI = "https://api.github.com"

// DevVersion é a versão de um binário compilado sem -ldflags (go build/go run).
const DevVersion = "dev"

// ErrNotFound indica que o repositório ou a release não existe ou exige token (repositório privado).
var ErrNotFound = errors.New("nenhuma release encontrada: ainda não foi publicada ou o repositório é privado (defina GITHUB_TOKEN ou GH_TOKEN, ou entre com `gh auth login`)")

// Asset é um arquivo de uma release.
type Asset struct {
	Name string `json:"name"`
	URL  string `json:"url"` // endpoint da API; serve o arquivo com Accept: application/octet-stream
}

// Release é a resposta de /releases/latest.
type Release struct {
	Tag    string  `json:"tag_name"`
	Assets []Asset `json:"assets"`
}

// Version devolve a tag sem o "v" inicial.
func (r Release) Version() string { return strings.TrimPrefix(r.Tag, "v") }

// Client fala com a API de releases do GitHub.
type Client struct {
	HTTP  *http.Client
	API   string // padrão DefaultAPI
	Repo  string // dono/nome
	Token string // opcional; necessário em repositório privado
}

func (c Client) api() string {
	if c.API != "" {
		return strings.TrimRight(c.API, "/")
	}
	return DefaultAPI
}

func (c Client) http() *http.Client {
	if c.HTTP != nil {
		return c.HTTP
	}
	return &http.Client{Timeout: 5 * time.Minute}
}

func (c Client) get(ctx context.Context, url, accept string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", accept)
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("User-Agent", "forge-cli")
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	resp, err := c.http().Do(req)
	if err != nil {
		return nil, err
	}
	switch {
	case resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusUnauthorized:
		_ = resp.Body.Close()
		return nil, ErrNotFound
	case resp.StatusCode >= 400:
		_ = resp.Body.Close()
		return nil, fmt.Errorf("GitHub respondeu HTTP %d", resp.StatusCode)
	}
	return resp, nil
}

// Latest consulta a release mais recente.
func (c Client) Latest(ctx context.Context) (Release, error) {
	resp, err := c.get(ctx, c.api()+"/repos/"+c.Repo+"/releases/latest", "application/vnd.github+json")
	if err != nil {
		return Release{}, err
	}
	defer func() { _ = resp.Body.Close() }()
	var r Release
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return Release{}, fmt.Errorf("resposta inválida do GitHub: %w", err)
	}
	if r.Tag == "" {
		return Release{}, ErrNotFound
	}
	return r, nil
}

func (c Client) download(ctx context.Context, a Asset) ([]byte, error) {
	resp, err := c.get(ctx, a.URL, "application/octet-stream")
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	return io.ReadAll(io.LimitReader(resp.Body, 200<<20))
}

// Compare compara versões "1.2.3" (com ou sem "v", pré-release ignorada): -1, 0 ou 1.
// Versões que não são numéricas (ex.: dev) contam como mais antigas que qualquer release.
func Compare(a, b string) int {
	pa, okA := parse(a)
	pb, okB := parse(b)
	switch {
	case !okA && !okB:
		return 0
	case !okA:
		return -1
	case !okB:
		return 1
	}
	for i := range pa {
		if pa[i] != pb[i] {
			if pa[i] < pb[i] {
				return -1
			}
			return 1
		}
	}
	return 0
}

func parse(v string) ([3]int, bool) {
	var out [3]int
	v = strings.TrimPrefix(strings.TrimSpace(v), "v")
	if i := strings.IndexAny(v, "-+"); i >= 0 {
		v = v[:i]
	}
	parts := strings.Split(v, ".")
	if v == "" || len(parts) > 3 {
		return out, false
	}
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 {
			return out, false
		}
		out[i] = n
	}
	return out, true
}

// AssetName é o nome do arquivo da release para o sistema (formato do .goreleaser.yaml).
func AssetName(version, goos, goarch string) string {
	ext := ".tar.gz"
	if goos == "windows" {
		ext = ".zip"
	}
	return fmt.Sprintf("forge_%s_%s_%s%s", strings.TrimPrefix(version, "v"), goos, goarch, ext)
}

// Result descreve o que a atualização fez (ou faria, em modo de checagem).
type Result struct {
	From, To string
	Newer    bool // existe versão mais nova
	Updated  bool // o binário foi trocado
}

// Update compara current com a última release e, se houver versão nova e check for false, baixa,
// confere o SHA-256 e troca o executável exe.
func (c Client) Update(ctx context.Context, current, exe, goos, goarch string, check bool) (Result, error) {
	rel, err := c.Latest(ctx)
	if err != nil {
		return Result{From: current}, err
	}
	res := Result{From: current, To: rel.Version(), Newer: Compare(current, rel.Version()) < 0}
	if !res.Newer || check {
		return res, nil
	}

	name := AssetName(rel.Version(), goos, goarch)
	var archive, sums *Asset
	for i := range rel.Assets {
		switch rel.Assets[i].Name {
		case name:
			archive = &rel.Assets[i]
		case "checksums.txt":
			sums = &rel.Assets[i]
		}
	}
	if archive == nil {
		return res, fmt.Errorf("a release %s não tem o arquivo %s", rel.Tag, name)
	}
	if sums == nil {
		return res, fmt.Errorf("a release %s não tem checksums.txt: não vou instalar sem conferir a integridade", rel.Tag)
	}

	data, err := c.download(ctx, *archive)
	if err != nil {
		return res, fmt.Errorf("baixando %s: %w", name, err)
	}
	sumFile, err := c.download(ctx, *sums)
	if err != nil {
		return res, fmt.Errorf("baixando checksums.txt: %w", err)
	}
	if err := verify(data, name, sumFile); err != nil {
		return res, err
	}
	bin, err := extract(data, name, goos)
	if err != nil {
		return res, err
	}
	if err := replaceExecutable(exe, bin); err != nil {
		return res, fmt.Errorf("trocando %s: %w", exe, err)
	}
	res.Updated = true
	return res, nil
}

func verify(data []byte, name string, checksums []byte) error {
	sum := sha256.Sum256(data)
	got := hex.EncodeToString(sum[:])
	for _, line := range strings.Split(string(checksums), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && strings.TrimPrefix(fields[1], "*") == name {
			if !strings.EqualFold(fields[0], got) {
				return fmt.Errorf("SHA-256 de %s não confere (esperado %s, obtido %s): download corrompido ou adulterado", name, fields[0], got)
			}
			return nil
		}
	}
	return fmt.Errorf("%s não consta em checksums.txt", name)
}

// extract tira o executável do .tar.gz ou .zip.
func extract(data []byte, name, goos string) ([]byte, error) {
	want := "forge"
	if goos == "windows" {
		want = "forge.exe"
	}
	if strings.HasSuffix(name, ".zip") {
		zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
		if err != nil {
			return nil, err
		}
		for _, f := range zr.File {
			if filepath.Base(f.Name) == want && !f.FileInfo().IsDir() {
				rc, err := f.Open()
				if err != nil {
					return nil, err
				}
				defer func() { _ = rc.Close() }()
				return io.ReadAll(io.LimitReader(rc, 200<<20))
			}
		}
		return nil, fmt.Errorf("%s não contém %s", name, want)
	}
	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer func() { _ = gz.Close() }()
	tr := tar.NewReader(gz)
	for {
		h, err := tr.Next()
		if errors.Is(err, io.EOF) {
			return nil, fmt.Errorf("%s não contém %s", name, want)
		}
		if err != nil {
			return nil, err
		}
		if h.Typeflag == tar.TypeReg && filepath.Base(h.Name) == want {
			return io.ReadAll(io.LimitReader(tr, 200<<20))
		}
	}
}

// replaceExecutable troca exe por bin. O executável em uso não pode ser sobrescrito no Windows, mas
// pode ser renomeado; o antigo (.old) é removido na próxima execução (Cleanup).
func replaceExecutable(exe string, bin []byte) error {
	mode := os.FileMode(0o755)
	if info, err := os.Stat(exe); err == nil {
		mode = info.Mode().Perm() | 0o111
	}
	tmp := exe + ".new"
	if err := os.WriteFile(tmp, bin, mode); err != nil {
		return err
	}
	old := exe + ".old"
	_ = os.Remove(old)
	if err := os.Rename(exe, old); err != nil && !errors.Is(err, os.ErrNotExist) {
		_ = os.Remove(tmp)
		return err
	}
	if err := os.Rename(tmp, exe); err != nil {
		_ = os.Rename(old, exe) // desfaz
		return err
	}
	_ = os.Remove(old) // falha no Windows enquanto o processo antigo roda
	return nil
}

// Cleanup apaga restos de uma atualização anterior (exe.old e exe.new).
func Cleanup(exe string) {
	_ = os.Remove(exe + ".old")
	_ = os.Remove(exe + ".new")
}
