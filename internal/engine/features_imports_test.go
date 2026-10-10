package engine

import "testing"

func TestPruneUnusedImports(t *testing.T) {
	in := `package p;

import a.b.Used;
import a.b.Gone;
import static a.b.Util.helper;
import static a.b.Util.unusedHelper;
import a.b.*;

class C {
  // forge:if auth=jwt
  Used u = helper();
  // forge:end
}
`
	want := `package p;

import a.b.Used;
import static a.b.Util.helper;
import a.b.*;

class C {
  Used u = helper();
}
`
	got, err := ProcessDirectives("C.java", in, sel("auth", "jwt"))
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}
