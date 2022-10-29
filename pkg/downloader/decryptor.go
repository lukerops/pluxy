package downloader

import (
	"crypto/aes"
	"crypto/cipher"
)

// https://stackoverflow.com/a/67627186
// https://gist.github.com/yingray/57fdc3264b1927ef0f984b533d63abab
func decryptAesCBC(data, key, iv []byte) ([]byte, error) {
	// Ajusta o tamanho do iv com
	// o tamanho do block esperado
	// pelo aes
	if len(iv) > aes.BlockSize {
		ivStart := len(iv) - aes.BlockSize
		iv = iv[ivStart:]
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	decoded := make([]byte, len(data))
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(decoded, data)

	return decoded, nil
}
