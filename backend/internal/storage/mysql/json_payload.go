package mysql

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

const (
	compressedJSONPayloadPrefix    = "gz64:"
	compressedJSONPayloadThreshold = 64 * 1024
)

func marshalJSONPayload(v any) (string, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	if len(raw) < compressedJSONPayloadThreshold {
		return string(raw), nil
	}

	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	if _, err := zw.Write(raw); err != nil {
		_ = zw.Close()
		return "", fmt.Errorf("compress json payload: %w", err)
	}
	if err := zw.Close(); err != nil {
		return "", fmt.Errorf("compress json payload: %w", err)
	}
	return compressedJSONPayloadPrefix + base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func unmarshalJSONPayload(payload string, v any) error {
	if !strings.HasPrefix(payload, compressedJSONPayloadPrefix) {
		return json.Unmarshal([]byte(payload), v)
	}

	encoded := strings.TrimPrefix(payload, compressedJSONPayloadPrefix)
	compressed, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return fmt.Errorf("decode compressed json payload: %w", err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return fmt.Errorf("open compressed json payload: %w", err)
	}
	defer zr.Close()
	raw, err := io.ReadAll(zr)
	if err != nil {
		return fmt.Errorf("read compressed json payload: %w", err)
	}
	return json.Unmarshal(raw, v)
}
