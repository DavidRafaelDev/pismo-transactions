package securelog

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestMaskJSONBody(t *testing.T) {
	t.Run("masks document_number at root", func(t *testing.T) {
		in := []byte(`{"document_number":"12345678900"}`)
		out, ok := MaskJSONBody(in)
		if !ok {
			t.Fatal("expected isJSON=true")
		}
		if !strings.Contains(string(out), `"document_number":"123******00"`) {
			t.Errorf("expected masked, got %s", out)
		}
	})

	t.Run("leaves non-sensitive fields untouched", func(t *testing.T) {
		in := []byte(`{"account_id":1,"document_number":"12345678900"}`)
		out, _ := MaskJSONBody(in)
		var decoded map[string]any
		if err := json.Unmarshal(out, &decoded); err != nil {
			t.Fatalf("output not JSON: %v", err)
		}
		if decoded["account_id"] != float64(1) {
			t.Errorf("account_id altered: %v", decoded["account_id"])
		}
		if decoded["document_number"] != "123******00" {
			t.Errorf("document_number not masked: %v", decoded["document_number"])
		}
	})

	t.Run("masks document_number nested in array", func(t *testing.T) {
		in := []byte(`[{"document_number":"11111111111"},{"document_number":"22222222222"}]`)
		out, _ := MaskJSONBody(in)
		if strings.Contains(string(out), "11111111111") {
			t.Errorf("raw document still present: %s", out)
		}
		if strings.Contains(string(out), "22222222222") {
			t.Errorf("raw document still present: %s", out)
		}
	})

	t.Run("masks document_number nested in nested object", func(t *testing.T) {
		in := []byte(`{"account":{"document_number":"12345678900","id":1}}`)
		out, _ := MaskJSONBody(in)
		if strings.Contains(string(out), "12345678900") {
			t.Errorf("raw document still present: %s", out)
		}
		if !strings.Contains(string(out), `"id":1`) {
			t.Errorf("sibling field lost: %s", out)
		}
	})

	t.Run("tolerates trailing whitespace (as emitted by json.NewEncoder)", func(t *testing.T) {
		in := []byte(`{"document_number":"12345678900"}` + "\n")
		_, ok := MaskJSONBody(in)
		if !ok {
			t.Error("expected isJSON=true even with trailing newline")
		}
	})

	t.Run("non-JSON body returned unchanged and flagged non-JSON", func(t *testing.T) {
		in := []byte("not-json-at-all")
		out, isJSON := MaskJSONBody(in)
		if isJSON {
			t.Error("expected isJSON=false")
		}
		if string(out) != string(in) {
			t.Errorf("non-JSON body altered: %s", out)
		}
	})

	t.Run("empty body returned unchanged", func(t *testing.T) {
		out, isJSON := MaskJSONBody(nil)
		if isJSON {
			t.Error("expected isJSON=false for empty")
		}
		if len(out) != 0 {
			t.Errorf("expected empty, got %s", out)
		}
	})
}
