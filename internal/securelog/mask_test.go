package securelog

import "testing"

func TestMaskDocument(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"cpf 11 digits", "12345678900", "123******00"},
		{"cnpj 14 digits", "12345678000199", "123*********99"},
		{"exactly 6 chars", "123456", "123*56"},
		{"short string", "abc", "***"},
		{"empty string", "", "***"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := MaskDocument(tc.input); got != tc.want {
				t.Errorf("MaskDocument(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}
