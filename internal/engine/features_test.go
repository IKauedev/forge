package engine

import (
	"strings"
	"testing"

	"github.com/ikauedev/forge/internal/template"
)

func sel(kv ...string) template.Selection {
	s := template.Selection{}
	for i := 0; i < len(kv); i += 2 {
		s[kv[i]] = kv[i+1]
	}
	return s
}

func TestProcessDirectives(t *testing.T) {
	tests := []struct {
		name, in, want string
		sel            template.Selection
	}{
		{
			name: "mantém ramo verdadeiro",
			in:   "a\n// forge:if auth!=none\nsec\n// forge:end\nb\n",
			want: "a\nsec\nb\n",
			sel:  sel("auth", "jwt"),
		},
		{
			name: "remove ramo falso",
			in:   "a\n// forge:if auth!=none\nsec\n// forge:end\nb\n",
			want: "a\nb\n",
			sel:  sel("auth", "none"),
		},
		{
			name: "else com alternativa comentada",
			in:   "# forge:if auth=cognito\n  #~ claim: groups\n# forge:else\n  claim: roles\n# forge:end\n",
			want: "  claim: groups\n",
			sel:  sel("auth", "cognito"),
		},
		{
			name: "else padrão",
			in:   "# forge:if auth=cognito\n  #~ claim: groups\n# forge:else\n  claim: roles\n# forge:end\n",
			want: "  claim: roles\n",
			sel:  sel("auth", "jwt"),
		},
		{
			name: "xml e lista de valores",
			in:   "<!-- forge:if http=feign,both -->\n<dep/>\n<!-- forge:end -->\n<!-- forge:if cache=redis -->\n<!--~ <redis/> -->\n<!-- forge:end -->\n",
			want: "<dep/>\n<redis/>\n",
			sel:  sel("http", "both", "cache", "redis"),
		},
		{
			name: "aninhado",
			in:   "// forge:if auth!=none\nA\n// forge:if cache=redis\nR\n// forge:end\nB\n// forge:end\n",
			want: "A\nB\n",
			sel:  sel("auth", "jwt", "cache", "none"),
		},
		{
			name: "aninhado dentro de ramo falso",
			in:   "// forge:if auth!=none\n// forge:if cache=redis\nR\n// forge:else\nX\n// forge:end\n// forge:end\n",
			want: "",
			sel:  sel("auth", "none", "cache", "none"),
		},
		{
			name: "condição com &&",
			in:   "// forge:if auth=jwt && cache=redis\nok\n// forge:end\n",
			want: "ok\n",
			sel:  sel("auth", "jwt", "cache", "redis"),
		},
		{
			name: "preserva CRLF",
			in:   "a\r\n//~ b\r\n",
			want: "a\r\nb\r\n",
			sel:  sel(),
		},
		{
			name: "sem diretivas passa intacto",
			in:   "x\ny",
			want: "x\ny",
			sel:  sel(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ProcessDirectives("f", tt.in, tt.sel)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestProcessDirectivesErrors(t *testing.T) {
	for name, in := range map[string]string{
		"if sem end":      "// forge:if auth=jwt\nx\n",
		"end sem if":      "// forge:end\n",
		"else sem if":     "// forge:else\n",
		"else duplicado":  "// forge:if auth=jwt\n// forge:else\n// forge:else\n// forge:end\n",
		"opção inválida":  "// forge:if nada=1\n// forge:end\n",
		"condição quebra": "// forge:if auth\n// forge:end\n",
	} {
		t.Run(name, func(t *testing.T) {
			_, err := ProcessDirectives("f.txt", in, sel("auth", "jwt"))
			if err == nil {
				t.Fatal("esperava erro")
			}
			if !strings.Contains(err.Error(), "f.txt") {
				t.Errorf("erro deveria citar o arquivo: %v", err)
			}
		})
	}
}
