// Package securelog provides helpers to redact sensitive fields before
// writing them to logs. The current implementation targets document_number
// (CPF/CNPJ), which is the only field flagged as sensitive in CLAUDE.md.
package securelog

import "strings"

// MaskDocument hides the middle digits of a document number, keeping the
// first 3 and last 2 visible for debuggability.
//
// Examples:
//
//	"12345678900"    -> "123******00"    (CPF, 11 digits)
//	"12345678000199" -> "123*********99" (CNPJ, 14 digits)
//	"abc"            -> "***"            (too short to mask meaningfully)
func MaskDocument(doc string) string {
	if len(doc) < 6 {
		return "***"
	}
	return doc[:3] + strings.Repeat("*", len(doc)-5) + doc[len(doc)-2:]
}
