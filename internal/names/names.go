// Package names normaliza o nome de um projeto em todos os formatos usados pelos templates.
package names

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// MaxKebabLen limita o nome kebab para caber nos 32 caracteres de ALB/target group da AWS,
// reservando 8 para sufixos como "-prod-tg".
const MaxKebabLen = 24

// Names contém o nome do projeto em cada formato.
type Names struct {
	Raw    string
	Kebab  string // payments-api
	Snake  string // payments_api
	Pascal string // PaymentsApi
	Flat   string // paymentsapi (segmento de pacote Java)
	Upper  string // PAYMENTS_API
	Title  string // Payments Api
}

var javaReserved = map[string]bool{
	"abstract": true, "assert": true, "boolean": true, "break": true, "byte": true, "case": true,
	"catch": true, "char": true, "class": true, "const": true, "continue": true, "default": true,
	"do": true, "double": true, "else": true, "enum": true, "extends": true, "final": true,
	"finally": true, "float": true, "for": true, "goto": true, "if": true, "implements": true,
	"import": true, "instanceof": true, "int": true, "interface": true, "long": true,
	"native": true, "new": true, "package": true, "private": true, "protected": true,
	"public": true, "return": true, "short": true, "static": true, "strictfp": true,
	"super": true, "switch": true, "synchronized": true, "this": true, "throw": true,
	"throws": true, "transient": true, "try": true, "void": true, "volatile": true,
	"while": true, "true": true, "false": true, "null": true, "_": true,
}

var packageRe = regexp.MustCompile(`^[a-z][a-z0-9_]*(\.[a-z][a-z0-9_]*)*$`)

// New valida e converte um nome livre ("Payments API", "payments-api", "PaymentsAPI"...).
func New(raw string) (Names, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return Names{}, errors.New("o nome do projeto não pode ser vazio")
	}
	for _, r := range s {
		if !isASCIILetter(r) && !isDigit(r) && r != ' ' && r != '-' && r != '_' {
			return Names{}, fmt.Errorf("caractere inválido %q: use letras ASCII, números, espaço, '-' ou '_'", r)
		}
	}
	if !isASCIILetter(rune(s[0])) {
		return Names{}, errors.New("o nome deve começar com uma letra")
	}

	words := splitWords(s)
	if len(words) == 0 {
		return Names{}, errors.New("o nome não contém palavras válidas")
	}

	titleWords := make([]string, len(words))
	pascalWords := make([]string, len(words))
	for i, w := range words {
		c := capitalize(w)
		titleWords[i], pascalWords[i] = c, c
	}
	n := Names{
		Raw:    raw,
		Kebab:  strings.Join(words, "-"),
		Snake:  strings.Join(words, "_"),
		Pascal: strings.Join(pascalWords, ""),
		Flat:   strings.Join(words, ""),
		Upper:  strings.ToUpper(strings.Join(words, "_")),
		Title:  strings.Join(titleWords, " "),
	}

	if javaReserved[n.Flat] {
		return Names{}, fmt.Errorf("%q é palavra reservada do Java e não pode ser um pacote", n.Flat)
	}
	if len(n.Kebab) > MaxKebabLen {
		return Names{}, fmt.Errorf("nome %q tem %d caracteres; o máximo é %d para respeitar os limites de nomes da AWS",
			n.Kebab, len(n.Kebab), MaxKebabLen)
	}
	return n, nil
}

// ValidatePackage valida o pacote base Java (ex.: com.empresa).
func ValidatePackage(pkg string) error {
	if !packageRe.MatchString(pkg) {
		return fmt.Errorf("pacote %q inválido: use segmentos minúsculos separados por '.' (ex.: com.empresa)", pkg)
	}
	for _, seg := range strings.Split(pkg, ".") {
		if javaReserved[seg] {
			return fmt.Errorf("pacote %q contém a palavra reservada %q", pkg, seg)
		}
	}
	return nil
}

// splitWords quebra por separadores e por fronteiras camelCase/acrônimo, em minúsculas.
func splitWords(s string) []string {
	runes := []rune(s)
	var words []string
	var cur []rune
	flush := func() {
		if len(cur) > 0 {
			words = append(words, strings.ToLower(string(cur)))
			cur = cur[:0]
		}
	}
	for i, r := range runes {
		if r == ' ' || r == '-' || r == '_' {
			flush()
			continue
		}
		if unicode.IsUpper(r) && len(cur) > 0 {
			prev := runes[i-1]
			nextLower := i+1 < len(runes) && unicode.IsLower(runes[i+1])
			if unicode.IsLower(prev) || isDigit(prev) || (unicode.IsUpper(prev) && nextLower) {
				flush()
			}
		}
		cur = append(cur, r)
	}
	flush()
	return words
}

func capitalize(w string) string {
	return strings.ToUpper(w[:1]) + w[1:]
}

func isASCIILetter(r rune) bool { return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') }
func isDigit(r rune) bool       { return r >= '0' && r <= '9' }
