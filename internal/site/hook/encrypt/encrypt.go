package encrypt

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"strings"
)

const (
	modeCBC = "cbc"
	modeCTR = "ctr"
	modeCFB = "cfb"
	modeOFB = "ofb"
)

type Encrypt struct {
	iv   []byte
	mode string
}

func normalizeMode(mode string) (string, error) {
	if mode == "" {
		return modeCBC, nil
	}

	mode = strings.ToLower(strings.TrimSpace(mode))
	switch mode {
	case modeCBC, modeCTR, modeCFB, modeOFB:
		return mode, nil
	default:
		return "", fmt.Errorf("unsupported encrypt mode %q", mode)
	}
}

func (e *Encrypt) deriveKey(key string) []byte {
	has := md5.Sum([]byte(key))
	return []byte(fmt.Sprintf("%x", has))
}

func (e *Encrypt) pkcs7Padding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

func (e *Encrypt) ivForBlockSize(blockSize int, ekey []byte) []byte {
	iv := e.iv
	if len(iv) != blockSize {
		iv = ekey[:blockSize]
	}
	return iv
}

func (e *Encrypt) Encrypt(plaintext, key string) (string, error) {
	data := []byte(plaintext)
	ekey := e.deriveKey(key)

	block, err := aes.NewCipher(ekey)
	if err != nil {
		return "", err
	}

	blockSize := block.BlockSize()
	iv := e.ivForBlockSize(blockSize, ekey)
	mode, err := normalizeMode(e.mode)
	if err != nil {
		return "", err
	}

	var encrypted []byte
	switch mode {
	case modeCBC:
		blockMode := cipher.NewCBCEncrypter(block, iv)
		encryptBytes := e.pkcs7Padding(data, blockSize)
		encrypted = make([]byte, len(encryptBytes))
		blockMode.CryptBlocks(encrypted, encryptBytes)
	case modeCTR:
		stream := cipher.NewCTR(block, iv)
		encrypted = make([]byte, len(data))
		stream.XORKeyStream(encrypted, data)
	case modeCFB:
		stream := cipher.NewCFBEncrypter(block, iv)
		encrypted = make([]byte, len(data))
		stream.XORKeyStream(encrypted, data)
	case modeOFB:
		stream := cipher.NewOFB(block, iv)
		encrypted = make([]byte, len(data))
		stream.XORKeyStream(encrypted, data)
	}

	return base64.StdEncoding.EncodeToString(encrypted), nil
}

func NewEncrypt(mode string) (*Encrypt, error) {
	mode, err := normalizeMode(mode)
	if err != nil {
		return nil, err
	}
	return &Encrypt{mode: mode}, nil
}
