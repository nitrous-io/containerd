package server

import (
	"crypto/rand"
	"encoding/hex"
	"io"
)

func randomString() string {
	id := make([]byte, 32)
	_, err := io.ReadFull(rand.Reader, id)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(id)
}
