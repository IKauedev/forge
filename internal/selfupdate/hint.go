package selfupdate

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// Hint é o resultado da checagem em segundo plano de versão nova.
type Hint struct {
	Latest string
	Newer  bool
}

type cacheFile struct {
	CheckedAt time.Time `json:"checked_at"`
	Latest    string    `json:"latest"`
}

// CheckCached diz se há versão nova, consultando a rede no máximo uma vez por ttl. Qualquer falha
// (sem rede, repositório privado sem token, cache ilegível) devolve ok=false: o aviso é só uma dica
// e nunca deve atrapalhar o uso.
func CheckCached(ctx context.Context, c Client, current, cachePath string, now time.Time, ttl time.Duration) (Hint, bool) {
	if Compare(current, "0.0.1") < 0 { // dev: não há com o que comparar
		return Hint{}, false
	}
	var cached cacheFile
	if b, err := os.ReadFile(cachePath); err == nil {
		_ = json.Unmarshal(b, &cached)
	}
	if cached.Latest != "" && now.Sub(cached.CheckedAt) < ttl {
		return Hint{Latest: cached.Latest, Newer: Compare(current, cached.Latest) < 0}, true
	}

	rel, err := c.Latest(ctx)
	if err != nil {
		return Hint{}, false
	}
	if b, err := json.Marshal(cacheFile{CheckedAt: now, Latest: rel.Version()}); err == nil {
		if os.MkdirAll(filepath.Dir(cachePath), 0o755) == nil {
			_ = os.WriteFile(cachePath, b, 0o644)
		}
	}
	return Hint{Latest: rel.Version(), Newer: Compare(current, rel.Version()) < 0}, true
}
