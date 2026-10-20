package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"io"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/blake2b"
)

const (
	Magic        = "envy1"
	SaltSize     = 16
	KeySize      = 32
	NonceSize    = 12
	argonTime    = 3
	argonMemory  = 64 * 1024 // KiB → 64 MiB
	argonThreads = 4
)

var (
	ErrInvalidKey       = errors.New("crypto: key must be 32 bytes")
	ErrUnknownMagic     = errors.New("crypto: unknown envelope magic")
	ErrCiphertextShort  = errors.New("crypto: ciphertext too short")
	ErrDecryptFailed    = errors.New("crypto: decryption failed")
	ErrVerifierMismatch = errors.New("crypto: passphrase verifier mismatch")
)

func GenerateSalt() ([]byte, error) {
	salt := make([]byte, SaltSize)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, fmt.Errorf("crypto: generate salt: %w", err)
	}
	return salt, nil
}

func DeriveKey(passphrase string, salt []byte) []byte {
	return argon2.IDKey([]byte(passphrase), salt, argonTime, argonMemory, argonThreads, KeySize)
}

func Verifier(key []byte) string {
	sum := blake2b.Sum256(key)
	return base64.StdEncoding.EncodeToString(sum[:])
}

func CheckVerifier(key []byte, stored string) bool {
	want, err := base64.StdEncoding.DecodeString(stored)
	if err != nil || len(want) != blake2b.Size256 {
		return false
	}
	got := blake2b.Sum256(key)
	return subtle.ConstantTimeCompare(got[:], want) == 1
}

func Seal(plaintext, key []byte) ([]byte, error) {
	if len(key) != KeySize {
		return nil, ErrInvalidKey
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, NonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("crypto: nonce: %w", err)
	}
	ct := gcm.Seal(nil, nonce, plaintext, nil)
	out := make([]byte, 0, len(Magic)+NonceSize+len(ct))
	out = append(out, Magic...)
	out = append(out, nonce...)
	out = append(out, ct...)
	return out, nil
}

func Open(blob, key []byte) ([]byte, error) {
	if len(key) != KeySize {
		return nil, ErrInvalidKey
	}
	if len(blob) < len(Magic)+NonceSize+gcmTagSize() {
		return nil, ErrCiphertextShort
	}
	if string(blob[:len(Magic)]) != Magic {
		return nil, ErrUnknownMagic
	}
	nonce := blob[len(Magic) : len(Magic)+NonceSize]
	ct := blob[len(Magic)+NonceSize:]
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	pt, err := gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		return nil, ErrDecryptFailed
	}
	return pt, nil
}

func gcmTagSize() int { return 16 }
