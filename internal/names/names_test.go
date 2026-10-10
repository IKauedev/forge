package names

import (
	"strings"
	"testing"
)

func TestNew_ValidInputs(t *testing.T) {
	want := Names{
		Kebab: "payments-api", Snake: "payments_api", Pascal: "PaymentsApi",
		Flat: "paymentsapi", Upper: "PAYMENTS_API", Title: "Payments Api",
	}
	for _, in := range []string{"Payments API", "payments-api", "PaymentsAPI", "payments_api", "  payments api  ", "PAYMENTS_API"} {
		t.Run(in, func(t *testing.T) {
			got, err := New(in)
			if err != nil {
				t.Fatalf("erro inesperado: %v", err)
			}
			got.Raw = ""
			if got != want {
				t.Errorf("got %+v, want %+v", got, want)
			}
		})
	}
}

func TestNew_OtherShapes(t *testing.T) {
	tests := []struct{ in, kebab, pascal string }{
		{"Payments", "payments", "Payments"},
		{"HTTPServer", "http-server", "HttpServer"},
		{"v2 api", "v2-api", "V2Api"},
		{"orders2Api", "orders2-api", "Orders2Api"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, err := New(tt.in)
			if err != nil {
				t.Fatal(err)
			}
			if got.Kebab != tt.kebab || got.Pascal != tt.pascal {
				t.Errorf("got kebab=%q pascal=%q, want %q %q", got.Kebab, got.Pascal, tt.kebab, tt.pascal)
			}
		})
	}
}

func TestNew_Invalid(t *testing.T) {
	tests := []struct{ name, in, errPart string }{
		{"vazio", "", "vazio"},
		{"só espaços", "   ", "vazio"},
		{"começa com número", "1api", "letra"},
		{"começa com hífen", "-api", "letra"},
		{"caractere especial", "my project!", "inválido"},
		{"acento", "pagamentos órfãos", "inválido"},
		{"palavra reservada", "class", "reservada"},
		{"reservada composta", "Inter Face", "interface"},
		{"muito longo", "this is a very long project name", "máximo"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.in)
			if err == nil {
				t.Fatal("esperava erro")
			}
			if !strings.Contains(err.Error(), tt.errPart) {
				t.Errorf("erro %q deveria conter %q", err, tt.errPart)
			}
		})
	}
}

func TestNew_AWSLimit(t *testing.T) {
	if _, err := New(strings.Repeat("a", MaxKebabLen)); err != nil {
		t.Errorf("tamanho exato do limite deveria passar: %v", err)
	}
	if _, err := New(strings.Repeat("a", MaxKebabLen+1)); err == nil {
		t.Error("acima do limite deveria falhar")
	}
}

func TestValidatePackage(t *testing.T) {
	for _, ok := range []string{"com.empresa", "com.example", "br.com.acme_corp", "org"} {
		if err := ValidatePackage(ok); err != nil {
			t.Errorf("%q deveria ser válido: %v", ok, err)
		}
	}
	for _, bad := range []string{"", "Com.empresa", "com..empresa", "com.1x", "com.empresa.", "com.class", "com-empresa"} {
		if err := ValidatePackage(bad); err == nil {
			t.Errorf("%q deveria ser inválido", bad)
		}
	}
}
