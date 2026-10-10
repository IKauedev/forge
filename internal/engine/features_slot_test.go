package engine

import (
	"strings"
	"testing"
)

const slotDoc = `before
# forge:slot greeting
reference
# forge:end
after
`

func TestSlotReplacedWhenSnippetExists(t *testing.T) {
	got, err := ProcessDirectivesWithSlots("f.txt", slotDoc, sel(), map[string]string{"greeting": "generated\n"})
	if err != nil {
		t.Fatal(err)
	}
	if want := "before\ngenerated\nafter\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestSlotKeepsReferenceWithoutSnippet(t *testing.T) {
	for _, slots := range []map[string]string{nil, {"other": "x"}} {
		got, err := ProcessDirectivesWithSlots("f.txt", slotDoc, sel(), slots)
		if err != nil {
			t.Fatal(err)
		}
		if want := "before\nreference\nafter\n"; got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}

func TestSlotAddsMissingNewline(t *testing.T) {
	got, _ := ProcessDirectivesWithSlots("f.txt", slotDoc, sel(), map[string]string{"greeting": "no-newline"})
	if want := "before\nno-newline\nafter\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestSlotInsideInactiveIfIsDropped(t *testing.T) {
	in := "// forge:if db!=none\n// forge:slot s\nref\n// forge:end\n// forge:end\nrest\n"
	got, err := ProcessDirectivesWithSlots("f.txt", in, sel("db", "none"), map[string]string{"s": "gen\n"})
	if err != nil {
		t.Fatal(err)
	}
	if got != "rest\n" {
		t.Errorf("slot dentro de if falso não pode vazar: %q", got)
	}
	got, _ = ProcessDirectivesWithSlots("f.txt", in, sel("db", "mysql"), map[string]string{"s": "gen\n"})
	if got != "gen\nrest\n" {
		t.Errorf("got %q", got)
	}
}

func TestSlotErrors(t *testing.T) {
	for name, in := range map[string]string{
		"sem nome":     "# forge:slot\nx\n# forge:end\n",
		"sem end":      "# forge:slot a\nx\n",
		"else no slot": "# forge:slot a\n# forge:else\n# forge:end\n",
	} {
		t.Run(name, func(t *testing.T) {
			_, err := ProcessDirectivesWithSlots("f.txt", in, sel(), nil)
			if err == nil || !strings.Contains(err.Error(), "f.txt") {
				t.Errorf("esperava erro citando o arquivo, veio %v", err)
			}
		})
	}
}

func TestOrConditions(t *testing.T) {
	in := "# forge:if db!=none||cache=redis\nyes\n# forge:end\n"
	for _, c := range []struct {
		db, cache string
		want      string
	}{{"mysql", "none", "yes\n"}, {"none", "redis", "yes\n"}, {"none", "none", ""}, {"none", "caffeine", ""}} {
		got, err := ProcessDirectives("f.txt", in, sel("db", c.db, "cache", c.cache))
		if err != nil || got != c.want {
			t.Errorf("db=%s cache=%s: got %q err %v, want %q", c.db, c.cache, got, err, c.want)
		}
	}
	// && tem precedência sobre ||
	in = "# forge:if db=none && cache=none || auth=jwt\nyes\n# forge:end\n"
	if got, _ := ProcessDirectives("f.txt", in, sel("db", "mysql", "cache", "none", "auth", "jwt")); got != "yes\n" {
		t.Errorf("auth=jwt sozinho deveria bastar: %q", got)
	}
	if got, _ := ProcessDirectives("f.txt", in, sel("db", "none", "cache", "redis", "auth", "none")); got != "" {
		t.Errorf("nenhum grupo verdadeiro: %q", got)
	}
}
