package md5

import (
	"crypto/md5"
	b64 "encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

const EmptyString = ""

// GenMd5HashAsBinary calculates the binary MD5 hash.
func GenMd5HashAsBinary(filePath string) ([]byte, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open the file '%s': %w", filePath, err)
	}
	defer f.Close()
	hasher := md5.New()
	if _, err := io.Copy(hasher, f); err == nil {
		return hasher.Sum(nil), nil
	} else {
		return nil, fmt.Errorf("failed to determine the binary MD5 hash of file '%s': %w", filePath, err)
	}
}

// GenMd5HashAsHex calculates the binary MD5 hash and encodes that value to a hex string.
func GenMd5HashAsHex(filePath string) (string, error) {
	if md5Hash, err := GenMd5HashAsBinary(filePath); err == nil {
		return hex.EncodeToString([]byte(md5Hash)), nil
	} else {
		return EmptyString, fmt.Errorf("failed to generate a MD5 hash as hex for '%s': %w", filePath, err)
	}
}

// GenMd5HashAsBase64 calculates the binary MD5 hash and encodes that value to a base64 string,
// which is how Azure Blob Storage records its Content-MD5 values.
func GenMd5HashAsBase64(filePath string) (string, error) {
	if md5Hash, err := GenMd5HashAsBinary(filePath); err == nil {
		return b64.StdEncoding.EncodeToString([]byte(md5Hash)), nil
	} else {
		return EmptyString, fmt.Errorf("failed to generate a MD5 hash as base64 for '%s': %w", filePath, err)
	}
}
