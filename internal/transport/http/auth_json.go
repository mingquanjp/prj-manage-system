package http

import (
	"bytes"
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"strconv"
	"unicode/utf16"
	"unicode/utf8"
)

const authBodyLimit = 8192

func hasAuthJSONContentType(r *http.Request) bool {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	return err == nil && mediaType == "application/json" && len(r.Header.Values("Content-Type")) == 1
}

func decodeAuthStringPair(body []byte, first, second string) (map[string]string, bool) {
	if !utf8.Valid(body) {
		return nil, false
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	if token, err := decoder.Token(); err != nil || token != json.Delim('{') {
		return nil, false
	}
	fields := make(map[string]string, 2)
	for decoder.More() {
		token, err := decoder.Token()
		if err != nil || (token != first && token != second) {
			return nil, false
		}
		key := token.(string)
		if _, duplicate := fields[key]; duplicate {
			return nil, false
		}
		var raw json.RawMessage
		if err := decoder.Decode(&raw); err != nil || raw[0] != '"' || !validJSONSurrogates(raw) {
			return nil, false
		}
		var value string
		if err := json.Unmarshal(raw, &value); err != nil {
			return nil, false
		}
		fields[key] = value
	}
	if token, err := decoder.Token(); err != nil || token != json.Delim('}') {
		return nil, false
	}
	if _, err := decoder.Token(); err != io.EOF {
		return nil, false
	}

	return fields, len(fields) == 2
}

// encoding/json đã kiểm cú pháp nhưng thay surrogate đơn bằng U+FFFD khi giải mã.
// Kiểm trên chuỗi JSON gốc để giữ nguyên ký tự U+FFFD được gửi hợp lệ.
func validJSONSurrogates(raw json.RawMessage) bool {
	for i := 1; i < len(raw)-1; i++ {
		if raw[i] != '\\' {
			continue
		}
		i++
		if raw[i] != 'u' {
			continue
		}
		unit, _ := strconv.ParseUint(string(raw[i+1:i+5]), 16, 16)
		i += 4
		if !utf16.IsSurrogate(rune(unit)) {
			continue
		}
		if unit >= 0xDC00 || i+6 >= len(raw)-1 || raw[i+1] != '\\' || raw[i+2] != 'u' {
			return false
		}
		low, _ := strconv.ParseUint(string(raw[i+3:i+7]), 16, 16)
		if low < 0xDC00 || low > 0xDFFF {
			return false
		}
		i += 6
	}
	return true
}
