package engine

import "testing"

func TestTidyJava(t *testing.T) {
	in := `class A {

  int a;

  // forge:if auth=none
  int b;

  // forge:end

  int c;

  // forge:if auth=none
  int d;
  // forge:end

}
`
	want := `class A {
  int a;

  int c;
}
`
	got, err := ProcessDirectives("A.java", in, sel("auth", "jwt"))
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestTidyJavaLeavesUntouchedFilesAlone(t *testing.T) {
	in := "class B {\n\n}\n"
	got, err := ProcessDirectives("B.java", in, sel())
	if err != nil {
		t.Fatal(err)
	}
	if got != in {
		t.Errorf("não deveria tocar em arquivo sem diretivas: %q", got)
	}
}
