package auth

import (
	cryptorand "crypto/rand"
)

func cryptoRandRead(b []byte) (int, error) {
	return cryptorand.Read(b)
}
