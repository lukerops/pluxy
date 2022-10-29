package downloader

import (
	"crypto/aes"
	"crypto/cipher"
)

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
