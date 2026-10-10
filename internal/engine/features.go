package engine

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/ikauedeveloper/forge/internal/template"
)

// Diretivas ficam em comentários, então o template de referência segue válido:
//
//	// forge:if auth!=none      #  forge:if http=feign,both      <!-- forge:if cache=redis -->
//	// forge:else
//	// forge:end
//
// Conteúdo que só existe numa opção diferente da referência é escrito comentado com "~"
// (//~ linha, #~ linha, <!--~ linha -->); o engine remove o marcador quando o ramo é escolhido.
// O google-java-format reescreve "//~" como "// ~", e os dois formatos são aceitos.
//
// Não coloque diretivas dentro do bloco de imports de arquivos .java: depois de processar, o
// engine poda sozinho os imports que ficaram sem uso.
var (
	directiveRe = regexp.MustCompile(`^\s*(?://|#|<!--)\s*forge:(if|else|end|slot)\b(.*?)\s*(?:-->)?\s*$`)
	lineMarkRe  = regexp.MustCompile(`^(\s*)(?://\s*|#)~ ?(.*)$`)
	xmlMarkRe   = regexp.MustCompile(`^(\s*)<!--~ ?(.*?)\s*-->\s*$`)
	importRe    = regexp.MustCompile(`^import\s+(static\s+)?([\w.]+)\.(\w+|\*);\s*$`)
)

type frame struct {
	parentActive bool
	cond         bool
	inElse       bool
	slot         bool // bloco forge:slot: o conteúdo interno é a referência
	replaced     bool // o slot foi trocado pelo trecho gerado
	line         int
}

func (f frame) active() bool {
	if f.slot {
		return f.parentActive && !f.replaced
	}
	if f.inElse {
		return f.parentActive && !f.cond
	}
	return f.parentActive && f.cond
}

// ProcessDirectives aplica as diretivas forge: de um arquivo de texto (sem slots gerados).
func ProcessDirectives(name, content string, sel template.Selection) (string, error) {
	return ProcessDirectivesWithSlots(name, content, sel, nil)
}

// ProcessDirectivesWithSlots também troca os blocos "forge:slot <nome>" pelos trechos em slots.
// Um slot sem trecho correspondente mantém o conteúdo de referência.
func ProcessDirectivesWithSlots(name, content string, sel template.Selection, slots map[string]string) (string, error) {
	lines := strings.SplitAfter(content, "\n")
	var out strings.Builder
	var stack []frame
	active := true
	changed := false

	for i, line := range lines {
		if line == "" {
			continue
		}
		if m := directiveRe.FindStringSubmatch(strings.TrimRight(line, "\r\n")); m != nil {
			switch m[1] {
			case "if":
				cond, err := template.Eval(strings.TrimSpace(m[2]), sel)
				if err != nil {
					return "", fmt.Errorf("%s:%d: %w", name, i+1, err)
				}
				stack = append(stack, frame{parentActive: active, cond: cond, line: i + 1})
			case "slot":
				slotName := strings.TrimSpace(m[2])
				if slotName == "" {
					return "", fmt.Errorf("%s:%d: forge:slot sem nome", name, i+1)
				}
				snippet, found := slots[slotName]
				replaced := found && active
				if replaced {
					out.WriteString(snippet)
					if !strings.HasSuffix(snippet, "\n") {
						out.WriteString("\n")
					}
				}
				stack = append(stack, frame{parentActive: active, slot: true, replaced: replaced, line: i + 1})
			case "else":
				if len(stack) == 0 || stack[len(stack)-1].inElse || stack[len(stack)-1].slot {
					return "", fmt.Errorf("%s:%d: forge:else sem forge:if", name, i+1)
				}
				stack[len(stack)-1].inElse = true
			case "end":
				if len(stack) == 0 {
					return "", fmt.Errorf("%s:%d: forge:end sem forge:if", name, i+1)
				}
				stack = stack[:len(stack)-1]
			}
			active = len(stack) == 0 || stack[len(stack)-1].active()
			changed = true
			continue
		}
		if active {
			un := unmark(line)
			changed = changed || un != line
			out.WriteString(un)
		}
	}
	if len(stack) > 0 {
		return "", fmt.Errorf("%s:%d: forge:if/forge:slot sem forge:end", name, stack[len(stack)-1].line)
	}
	if changed && strings.HasSuffix(name, ".java") {
		return pruneUnusedImports(tidyJava(out.String())), nil
	}
	return out.String(), nil
}

// tidyJava remove as linhas em branco que a remoção de blocos deixa e que o google-java-format
// rejeitaria: duplicadas, logo após "{" ou logo antes de "}".
func tidyJava(src string) string {
	lines := strings.SplitAfter(src, "\n")
	var out []string
	for i, line := range lines {
		if strings.TrimSpace(line) != "" {
			out = append(out, line)
			continue
		}
		if len(out) == 0 || strings.TrimSpace(out[len(out)-1]) == "" ||
			strings.HasSuffix(strings.TrimSpace(out[len(out)-1]), "{") {
			continue
		}
		next := ""
		for _, l := range lines[i+1:] {
			if strings.TrimSpace(l) != "" {
				next = strings.TrimSpace(l)
				break
			}
		}
		if strings.HasPrefix(next, "}") {
			continue
		}
		out = append(out, line)
	}
	return strings.Join(out, "")
}

// pruneUnusedImports remove imports cujo nome simples não aparece no resto do arquivo. Evita que
// um bloco removido deixe import de classe ou biblioteca que não existe mais.
func pruneUnusedImports(src string) string {
	lines := strings.SplitAfter(src, "\n")
	var body strings.Builder
	for _, l := range lines {
		if !importRe.MatchString(strings.TrimRight(l, "\r\n")) {
			body.WriteString(l)
		}
	}
	text := body.String()

	var out strings.Builder
	for _, l := range lines {
		m := importRe.FindStringSubmatch(strings.TrimRight(l, "\r\n"))
		if m != nil && m[3] != "*" && !usesName(text, m[3]) {
			continue
		}
		out.WriteString(l)
	}
	return out.String()
}

func usesName(text, name string) bool {
	return regexp.MustCompile(`\b` + regexp.QuoteMeta(name) + `\b`).MatchString(text)
}

func unmark(line string) string {
	trimmed, eol := splitEOL(line)
	if m := lineMarkRe.FindStringSubmatch(trimmed); m != nil {
		return m[1] + m[2] + eol
	}
	if m := xmlMarkRe.FindStringSubmatch(trimmed); m != nil {
		return m[1] + m[2] + eol
	}
	return line
}

func splitEOL(line string) (string, string) {
	body := strings.TrimRight(line, "\r\n")
	return body, line[len(body):]
}
