package engine

import "testing"

// O google-java-format reescreve "//~" como "// ~"; os dois precisam funcionar.
func TestMarkerAcceptsFormatterStyle(t *testing.T) {
	in := "// forge:if cache=redis\n  // ~ int a;\n  //~ int b;\n// forge:end\n"
	got, err := ProcessDirectives("A.txt", in, sel("cache", "redis"))
	if err != nil {
		t.Fatal(err)
	}
	if want := "  int a;\n  int b;\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
