package orm

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	filestoreMu   sync.RWMutex
	filestoreRoot string
)

// InitFilestore sets the on-disk attachment directory (created if missing).
func InitFilestore(root string) error {
	root = strings.TrimSpace(root)
	if root == "" {
		root = filepath.Join(os.TempDir(), "sumeru-filestore")
	}
	if err := os.MkdirAll(root, 0o750); err != nil {
		return fmt.Errorf("filestore mkdir: %w", err)
	}
	filestoreMu.Lock()
	filestoreRoot = root
	filestoreMu.Unlock()
	return nil
}

func filestorePath(key string) (string, error) {
	filestoreMu.RLock()
	root := filestoreRoot
	filestoreMu.RUnlock()
	if root == "" {
		if err := InitFilestore(""); err != nil {
			return "", err
		}
		filestoreMu.RLock()
		root = filestoreRoot
		filestoreMu.RUnlock()
	}
	if key == "" {
		return "", fmt.Errorf("filestore: empty key")
	}
	hash := sha256.Sum256([]byte(key))
	hexKey := hex.EncodeToString(hash[:])
	return filepath.Join(root, hexKey[:2], hexKey[2:4], hexKey), nil
}

// StoreAttachment writes blob bytes and returns the stored filename key.
func StoreAttachment(_ context.Context, name string, data []byte) (storeFname string, size int64, err error) {
	if len(data) == 0 {
		return "", 0, fmt.Errorf("filestore: empty data")
	}
	key := strings.TrimSpace(name)
	if key == "" {
		key = hex.EncodeToString(sha256Sum(data))
	}
	path, err := filestorePath(key)
	if err != nil {
		return "", 0, err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return "", 0, err
	}
	if err := os.WriteFile(path, data, 0o640); err != nil {
		return "", 0, fmt.Errorf("filestore write: %w", err)
	}
	return key, int64(len(data)), nil
}

// ReadAttachment loads blob bytes by store key.
func ReadAttachment(_ context.Context, storeFname string) ([]byte, error) {
	path, err := filestorePath(storeFname)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("filestore read: %w", err)
	}
	return data, nil
}

// OpenAttachment returns a ReadCloser for streaming reads.
func OpenAttachment(_ context.Context, storeFname string) (io.ReadCloser, error) {
	path, err := filestorePath(storeFname)
	if err != nil {
		return nil, err
	}
	return os.Open(path)
}

func sha256Sum(data []byte) []byte {
	sum := sha256.Sum256(data)
	return sum[:]
}
