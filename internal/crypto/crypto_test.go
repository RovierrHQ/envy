package crypto

import (
	"bytes"
	"encoding/base64"
	"testing"
)

func TestDeriveKeyDeterministic(t *testing.T) {
	salt := bytes.Repeat([]byte{0x42}, SaltSize)
	k1 := DeriveKey("correct horse battery staple", salt)
	k2 := DeriveKey("correct horse battery staple", salt)
	if !bytes.Equal(k1, k2) {
		t.Fatal("derive not deterministic")
	}
	if len(k1) != KeySize {
		t.Fatalf("key len=%d want %d", len(k1), KeySize)
	}
	k3 := DeriveKey("wrong passphrase", salt)
	if bytes.Equal(k1, k3) {
		t.Fatal("different passphrases produced same key")
	}
}

func TestVerifierRoundTrip(t *testing.T) {
	salt, err := GenerateSalt()
	if err != nil {
		t.Fatal(err)
	}
	key := DeriveKey("test-pass", salt)
	v := Verifier(key)
	if !CheckVerifier(key, v) {
		t.Fatal("verifier should match")
	}
	if CheckVerifier(key, "not-valid-base64!!!") {
		t.Fatal("invalid verifier should fail")
	}
	other := DeriveKey("other", salt)
	if CheckVerifier(other, v) {
		t.Fatal("wrong key should fail verifier")
	}
	_, err = base64.StdEncoding.DecodeString(v)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSealOpenRoundTrip(t *testing.T) {
	key := DeriveKey("secret", bytes.Repeat([]byte{1}, SaltSize))
	pt := []byte("FOO=bar\nBAZ=qux\n")
	blob, err := Seal(pt, key)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(blob, []byte(Magic)) {
		t.Fatalf("missing magic: %q", blob[:min(8, len(blob))])
	}
	got, err := Open(blob, key)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, pt) {
		t.Fatalf("got %q want %q", got, pt)
	}
}

func TestOpenWrongKey(t *testing.T) {
	salt := bytes.Repeat([]byte{2}, SaltSize)
	key := DeriveKey("a", salt)
	blob, err := Seal([]byte("x=1"), key)
	if err != nil {
		t.Fatal(err)
	}
	_, err = Open(blob, DeriveKey("b", salt))
	if err != ErrDecryptFailed {
		t.Fatalf("want ErrDecryptFailed got %v", err)
	}
}

func TestOpenTampered(t *testing.T) {
	key := DeriveKey("a", bytes.Repeat([]byte{3}, SaltSize))
	blob, err := Seal([]byte("x=1"), key)
	if err != nil {
		t.Fatal(err)
	}
	blob[len(blob)-1] ^= 0xff
	_, err = Open(blob, key)
	if err != ErrDecryptFailed {
		t.Fatalf("want ErrDecryptFailed got %v", err)
	}
}

func TestOpenTruncated(t *testing.T) {
	key := make([]byte, KeySize)
	_, err := Open([]byte("envy1"), key)
	if err != ErrCiphertextShort {
		t.Fatalf("want ErrCiphertextShort got %v", err)
	}
}

func TestOpenUnknownMagic(t *testing.T) {
	key := make([]byte, KeySize)
	blob := append([]byte("envy2"), make([]byte, NonceSize+16)...)
	_, err := Open(blob, key)
	if err != ErrUnknownMagic {
		t.Fatalf("want ErrUnknownMagic got %v", err)
	}
}

func TestSealInvalidKey(t *testing.T) {
	_, err := Seal([]byte("x"), []byte("short"))
	if err != ErrInvalidKey {
		t.Fatalf("want ErrInvalidKey got %v", err)
	}
}

func TestGenerateSalt(t *testing.T) {
	a, err := GenerateSalt()
	if err != nil {
		t.Fatal(err)
	}
	b, err := GenerateSalt()
	if err != nil {
		t.Fatal(err)
	}
	if len(a) != SaltSize || len(b) != SaltSize {
		t.Fatal("bad salt size")
	}
	if bytes.Equal(a, b) {
		t.Fatal("salts should differ")
	}
}
