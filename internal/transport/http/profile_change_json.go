package http

import (
	"bytes"
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"

	pc "ai-staffing/internal/service/profilechange"
)

const profileChangeBodyLimit = 65536

var profileChangeKeyPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-4[0-9a-fA-F]{3}-[89aAbB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`)

func hasProfileChangeJSONContentType(r *http.Request) bool {
	if len(r.Header.Values("Content-Type")) != 1 || len(r.Header.Values("Content-Encoding")) > 1 {
		return false
	}
	encoding := r.Header.Get("Content-Encoding")
	if encoding != "" && !strings.EqualFold(encoding, "identity") {
		return false
	}
	media, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || media != "application/json" {
		return false
	}
	for key, value := range params {
		if key != "charset" || !strings.EqualFold(value, "utf-8") {
			return false
		}
	}
	return true
}

// Kiểm chuỗi JSON gốc trước khi encoding/json có thể thay surrogate đơn bằng U+FFFD.
// json.Valid bảo đảm biên escape hợp lệ cho validJSONSurrogates.
func validProfileChangeJSON(body []byte) bool {
	if !utf8.Valid(body) || !json.Valid(body) {
		return false
	}
	for i := 0; i < len(body); i++ {
		if body[i] != '"' {
			continue
		}
		start := i
		for i++; i < len(body); i++ {
			if body[i] == '\\' {
				i++
				continue
			}
			if body[i] == '"' {
				break
			}
		}
		raw := body[start : i+1]
		var value string
		if !validJSONSurrogates(raw) || json.Unmarshal(raw, &value) != nil || strings.ContainsRune(value, 0) {
			return false
		}
	}
	return true
}

func profileChangeObject(body []byte, allowed ...string) (map[string]json.RawMessage, error) {
	d := json.NewDecoder(bytes.NewReader(body))
	if token, err := d.Token(); err != nil || token != json.Delim('{') {
		return nil, pc.ErrValidation
	}
	fields := map[string]json.RawMessage{}
	for d.More() {
		token, err := d.Token()
		key, ok := token.(string)
		if err != nil || !ok {
			return nil, pc.ErrValidation
		}
		found := false
		for _, name := range allowed {
			found = found || key == name
		}
		if _, duplicate := fields[key]; !found || duplicate {
			return nil, pc.ErrValidation
		}
		var raw json.RawMessage
		if d.Decode(&raw) != nil {
			return nil, pc.ErrValidation
		}
		fields[key] = raw
	}
	if token, err := d.Token(); err != nil || token != json.Delim('}') {
		return nil, pc.ErrValidation
	}
	if _, err := d.Token(); err != io.EOF {
		return nil, pc.ErrValidation
	}
	return fields, nil
}

func decodeProfileChangeValues(raw []byte) (pc.Values, error) {
	fields, err := profileChangeObject(raw, "fullname", "phone", "address")
	if err != nil || len(fields) < 1 {
		return nil, pc.ErrValidation
	}
	values := pc.Values{}
	for key, raw := range fields {
		var value *string
		if json.Unmarshal(raw, &value) != nil {
			return nil, pc.ErrValidation
		}
		values[key] = value
	}
	return values, nil
}

func decodeProfileChangeBody(body []byte, operation string) (pc.SubmitInput, string, error) {
	var input pc.SubmitInput
	if !validProfileChangeJSON(body) {
		return input, "", pc.ErrValidation
	}
	switch operation {
	case "submit":
		fields, err := profileChangeObject(body, "expected_before_values", "proposed_values", "explanation")
		if err != nil {
			return input, "", err
		}
		input.Before, err = decodeProfileChangeValues(fields["expected_before_values"])
		if err != nil {
			return input, "", err
		}
		input.Proposed, err = decodeProfileChangeValues(fields["proposed_values"])
		if err != nil {
			return input, "", err
		}
		if raw, present := fields["explanation"]; present && json.Unmarshal(raw, &input.Explanation) != nil {
			return input, "", pc.ErrValidation
		}
		return input, "", nil
	case "approve":
		_, err := profileChangeObject(body)
		return input, "", err
	case "reject":
		fields, err := profileChangeObject(body, "reason")
		var reason *string
		if err != nil || len(fields) != 1 || json.Unmarshal(fields["reason"], &reason) != nil || reason == nil {
			return input, "", pc.ErrValidation
		}
		return input, *reason, nil
	}
	return input, "", pc.ErrValidation
}

func parseProfileChangeQuery(raw string) (pc.ListQuery, error) {
	q := pc.ListQuery{Scope: "self", Page: 1, PageSize: 20}
	values, err := url.ParseQuery(raw)
	if err != nil {
		return q, pc.ErrValidation
	}
	for key, v := range values {
		if len(v) != 1 {
			return q, pc.ErrValidation
		}
		switch key {
		case "scope":
			if v[0] != "self" && v[0] != "review" {
				return q, pc.ErrValidation
			}
			q.Scope = v[0]
		case "status":
			if v[0] != "PENDING" && v[0] != "APPROVED" && v[0] != "REJECTED" {
				return q, pc.ErrValidation
			}
			q.Status = v[0]
		case "page", "page_size":
			n, err := parseCatalogID(v[0])
			if err != nil || (key == "page_size" && n > 100) {
				return q, pc.ErrValidation
			}
			if key == "page" {
				q.Page = int64(n)
			} else {
				q.PageSize = n
			}
		default:
			return q, pc.ErrValidation
		}
	}
	return q, nil
}
