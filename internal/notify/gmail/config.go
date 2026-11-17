package gmail

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type credentials struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RefreshToken string `json:"refresh_token"`
	From         string `json:"from"`
}

func (c credentials) valid() bool {
	return nonblank(c.ClientID) && nonblank(c.ClientSecret) && nonblank(c.RefreshToken) && validAddress(c.From)
}

func nonblank(value string) bool {
	return value != "" && strings.TrimSpace(value) == value && !strings.ContainsAny(value, "\r\n\x00")
}

func credentialPath() string {
	return os.Getenv("GMAIL_OAUTH_CREDENTIALS_FILE")
}

func decodeLimitedJSON(r io.Reader, value any) error {
	const max = 64 << 10
	data, err := io.ReadAll(io.LimitReader(r, max+1))
	if err != nil || len(data) > max || json.Unmarshal(data, value) != nil {
		return ErrConfig
	}
	return nil
}

func readPrivateJSON(path string, value any) error {
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0077 != 0 {
		return ErrConfig
	}
	f, err := os.Open(path)
	if err != nil {
		return ErrConfig
	}
	defer f.Close()
	opened, err := f.Stat()
	if err != nil || !os.SameFile(info, opened) || !opened.Mode().IsRegular() || opened.Mode().Perm()&0077 != 0 {
		return ErrConfig
	}
	return decodeLimitedJSON(f, value)
}

func writeCredentials(path string, value credentials) error {
	if !filepath.IsAbs(path) || !value.valid() {
		return ErrConfig
	}
	dir := filepath.Dir(path)
	if os.MkdirAll(dir, 0700) != nil {
		return ErrConfig
	}
	info, err := os.Lstat(dir)
	if err != nil || !info.IsDir() || info.Mode().Perm()&0077 != 0 {
		return ErrConfig
	}
	if existing, err := os.Lstat(path); err == nil {
		if !existing.Mode().IsRegular() || existing.Mode().Perm()&0077 != 0 {
			return ErrConfig
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return ErrConfig
	}
	temp, err := os.CreateTemp(dir, ".gmail-oauth-*")
	if err != nil {
		return ErrConfig
	}
	defer os.Remove(temp.Name())
	defer temp.Close()
	if json.NewEncoder(temp).Encode(value) != nil || temp.Sync() != nil || temp.Close() != nil || os.Rename(temp.Name(), path) != nil {
		return ErrConfig
	}
	return nil
}
