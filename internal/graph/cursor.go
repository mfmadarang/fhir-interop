package graph

import "encoding/base64"

// encodes a patient ID into an opaque cursor
func encodeCursor(id string) string {
	return base64.StdEncoding.EncodeToString([]byte(id))
}

// decodes an opaque cursor back into a patient ID
func decodeCursor(cursor string) (string, error) {
	b, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
