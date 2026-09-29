package domain

import (
	"errors"
	"testing"
	"time"
)

func TestNewAccount(t *testing.T) {
	tests := []struct {
		name    string
		doc     string
		wantErr error
	}{
		{"cpf ok (11 digits)", "12345678900", nil},
		{"cnpj ok (14 digits)", "12345678000199", nil},
		{"empty", "", ErrInvalidDocumentNumber},
		{"too short", "123", ErrInvalidDocumentNumber},
		{"contains letter", "1234567890A", ErrInvalidDocumentNumber},
		{"12 digits (invalid length)", "123456789012", ErrInvalidDocumentNumber},
		{"13 digits (invalid length)", "1234567890123", ErrInvalidDocumentNumber},
		{"15 digits (too long)", "123456789001234", ErrInvalidDocumentNumber},
		{"whitespace inside", "123 4567 8900", ErrInvalidDocumentNumber},
		{"formatted with dots and dashes", "123.456.789-00", ErrInvalidDocumentNumber},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			acc, err := NewAccount(tc.doc)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("err: want %v, got %v", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if acc.DocumentNumber != tc.doc {
				t.Errorf("DocumentNumber: want %q, got %q", tc.doc, acc.DocumentNumber)
			}
			if time.Since(acc.CreatedAt) > time.Second {
				t.Errorf("CreatedAt too old: %v", acc.CreatedAt)
			}
			if acc.CreatedAt.Location() != time.UTC {
				t.Errorf("CreatedAt not UTC: %v", acc.CreatedAt.Location())
			}
			if acc.ID != 0 {
				t.Errorf("ID should be zero before persist, got %d", acc.ID)
			}
		})
	}
}
