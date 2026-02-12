package x

import (
	"crypto/ecdsa"
	"os"
	"path/filepath"
	"testing"
)

func TestMakeEllipticPrivateKeyPEM(t *testing.T) {
	pem, err := MakeEllipticPrivateKeyPEM()
	if err != nil {
		t.Fatalf("MakeEllipticPrivateKeyPEM() error = %v", err)
	}
	if len(pem) == 0 {
		t.Error("MakeEllipticPrivateKeyPEM() returned empty PEM")
	}
	if string(pem[:27]) != "-----BEGIN EC PRIVATE KEY--" {
		t.Errorf("MakeEllipticPrivateKeyPEM() does not start with expected header")
	}
}

func TestParsePrivateKeyPEM_ECDSA(t *testing.T) {
	pem, err := MakeEllipticPrivateKeyPEM()
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	key, err := ParsePrivateKeyPEM(pem)
	if err != nil {
		t.Fatalf("ParsePrivateKeyPEM() error = %v", err)
	}

	if _, ok := key.(*ecdsa.PrivateKey); !ok {
		t.Errorf("ParsePrivateKeyPEM() returned %T, want *ecdsa.PrivateKey", key)
	}
}

func TestParsePrivateKeyPEM_Invalid(t *testing.T) {
	_, err := ParsePrivateKeyPEM([]byte("not a valid pem"))
	if err == nil {
		t.Error("ParsePrivateKeyPEM() should return error for invalid PEM")
	}
}

func TestParsePrivateKeyPEM_Empty(t *testing.T) {
	_, err := ParsePrivateKeyPEM([]byte{})
	if err == nil {
		t.Error("ParsePrivateKeyPEM() should return error for empty data")
	}
}

func TestWriteKey(t *testing.T) {
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "subdir", "test.key")

	data := []byte("test key data")
	err := WriteKey(keyPath, data)
	if err != nil {
		t.Fatalf("WriteKey() error = %v", err)
	}

	read, err := os.ReadFile(keyPath)
	if err != nil {
		t.Fatalf("Failed to read written key: %v", err)
	}
	if string(read) != string(data) {
		t.Errorf("WriteKey() wrote %q, want %q", read, data)
	}

	info, err := os.Stat(keyPath)
	if err != nil {
		t.Fatalf("Failed to stat key file: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("Key file permissions = %o, want 0600", info.Mode().Perm())
	}
}

func TestLoadOrGenerateKeyFile_Generate(t *testing.T) {
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "new.key")

	data, wasGenerated, err := LoadOrGenerateKeyFile(keyPath)
	if err != nil {
		t.Fatalf("LoadOrGenerateKeyFile() error = %v", err)
	}
	if !wasGenerated {
		t.Error("LoadOrGenerateKeyFile() wasGenerated = false, want true")
	}
	if len(data) == 0 {
		t.Error("LoadOrGenerateKeyFile() returned empty data")
	}

	_, err = ParsePrivateKeyPEM(data)
	if err != nil {
		t.Errorf("Generated key is not valid: %v", err)
	}
}

func TestLoadOrGenerateKeyFile_Load(t *testing.T) {
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "existing.key")

	originalData, err := MakeEllipticPrivateKeyPEM()
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}
	if err := WriteKey(keyPath, originalData); err != nil {
		t.Fatalf("Failed to write key: %v", err)
	}

	data, wasGenerated, err := LoadOrGenerateKeyFile(keyPath)
	if err != nil {
		t.Fatalf("LoadOrGenerateKeyFile() error = %v", err)
	}
	if wasGenerated {
		t.Error("LoadOrGenerateKeyFile() wasGenerated = true, want false")
	}
	if string(data) != string(originalData) {
		t.Error("LoadOrGenerateKeyFile() returned different data than original")
	}
}

func TestMarshalPrivateKeyToPEM_ECDSA(t *testing.T) {
	keyPEM, err := MakeEllipticPrivateKeyPEM()
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}
	key, err := ParsePrivateKeyPEM(keyPEM)
	if err != nil {
		t.Fatalf("Failed to parse key: %v", err)
	}

	pem, err := MarshalPrivateKeyToPEM(key)
	if err != nil {
		t.Fatalf("MarshalPrivateKeyToPEM() error = %v", err)
	}
	if len(pem) == 0 {
		t.Error("MarshalPrivateKeyToPEM() returned empty PEM")
	}
}

func TestMarshalPrivateKeyToPEM_Unsupported(t *testing.T) {
	_, err := MarshalPrivateKeyToPEM("not a key")
	if err == nil {
		t.Error("MarshalPrivateKeyToPEM() should return error for unsupported type")
	}
}

func TestPrivateKeyFromFile(t *testing.T) {
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "test.key")

	keyPEM, _ := MakeEllipticPrivateKeyPEM()
	WriteKey(keyPath, keyPEM)

	key, err := PrivateKeyFromFile(keyPath)
	if err != nil {
		t.Fatalf("PrivateKeyFromFile() error = %v", err)
	}
	if _, ok := key.(*ecdsa.PrivateKey); !ok {
		t.Errorf("PrivateKeyFromFile() returned %T, want *ecdsa.PrivateKey", key)
	}
}

func TestPrivateKeyFromFile_NotFound(t *testing.T) {
	_, err := PrivateKeyFromFile("/nonexistent/path/key.pem")
	if err == nil {
		t.Error("PrivateKeyFromFile() should return error for nonexistent file")
	}
}

func TestPublicKeysFromFile(t *testing.T) {
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "test.key")

	keyPEM, _ := MakeEllipticPrivateKeyPEM()
	WriteKey(keyPath, keyPEM)

	keys, err := PublicKeysFromFile(keyPath)
	if err != nil {
		t.Fatalf("PublicKeysFromFile() error = %v", err)
	}
	if len(keys) != 1 {
		t.Errorf("PublicKeysFromFile() returned %d keys, want 1", len(keys))
	}
	if _, ok := keys[0].(*ecdsa.PublicKey); !ok {
		t.Errorf("PublicKeysFromFile() returned %T, want *ecdsa.PublicKey", keys[0])
	}
}

func TestParsePublicKeysPEM_Invalid(t *testing.T) {
	_, err := ParsePublicKeysPEM([]byte("not valid pem"))
	if err == nil {
		t.Error("ParsePublicKeysPEM() should return error for invalid PEM")
	}
}
