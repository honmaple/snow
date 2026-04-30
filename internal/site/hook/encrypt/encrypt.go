package encrypt

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/base64"
	"fmt"
)

type (
	Encrypt struct {
		iv []byte
	}
)

func (e *Encrypt) deriveKey(key string) []byte {
	has := md5.Sum([]byte(key))
	return []byte(fmt.Sprintf("%x", has))
}

func (e *Encrypt) pkcs7Padding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

func (e *Encrypt) Encrypt(plaintext, key string) (string, error) {
	data := []byte(plaintext)
	ekey := e.deriveKey(key)

	block, err := aes.NewCipher(ekey)
	if err != nil {
		return "", err
	}
	blockSize := block.BlockSize()
	iv := e.iv
	if len(iv) != blockSize {
		iv = ekey[:blockSize]
	}
	blockMode := cipher.NewCBCEncrypter(block, iv)

	encryptBytes := e.pkcs7Padding(data, blockSize)
	encrypted := make([]byte, len(encryptBytes))
	blockMode.CryptBlocks(encrypted, encryptBytes)
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

func NewEncrypt() *Encrypt {
	return &Encrypt{}
}
