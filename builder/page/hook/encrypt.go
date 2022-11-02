package hook

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"errors"

	"github.com/honmaple/snow/builder/hook"
	"github.com/honmaple/snow/builder/page"
	"github.com/honmaple/snow/builder/theme"
	"github.com/honmaple/snow/config"
)

type Encrypt struct {
	hook.BaseHook
	conf config.Config
}

var (
	errInvalidBlockSize    = errors.New("invalid blocksize")
	errInvalidPKCS7Data    = errors.New("invalid PKCS7 data (empty or not padded)")
	errInvalidPKCS7Padding = errors.New("invalid padding on input")
)

func pkcs7Padding(plainText []byte, blockSize int) []byte {
	padding := blockSize - len(plainText)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(plainText, padText...)
}

func pkcs7Unpad(b []byte, blocksize int) ([]byte, error) {
	if blocksize <= 0 {
		return nil, errInvalidBlockSize
	}
	if b == nil || len(b) == 0 {
		return nil, errInvalidPKCS7Data
	}
	if len(b)%blocksize != 0 {
		return nil, errInvalidPKCS7Padding
	}
	c := b[len(b)-1]
	n := int(c)
	if n == 0 || n > len(b) {
		return nil, errInvalidPKCS7Padding
	}
	for i := 0; i < n; i++ {
		if b[len(b)-n+i] != c {
			return nil, errInvalidPKCS7Padding
		}
	}
	return b[:len(b)-n], nil
}

func (e *Encrypt) encrypt(plainText []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()

	plainText = pkcs7Padding(plainText, blockSize)

	cipherText := make([]byte, len(plainText))
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	blockMode.CryptBlocks(cipherText, plainText)
	return cipherText, nil
}

func (e *Encrypt) decrypt(cipherText []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()

	plainText := make([]byte, len(cipherText))
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	blockMode.CryptBlocks(plainText, cipherText)

	return pkcs7Unpad(plainText, blockSize)
}

func (e *Encrypt) AfterPageParse(meta map[string]string, page *page.Page) *page.Page {
	password := meta["password"]
	if password == "" {
		return page
	}
	c, err := e.encrypt([]byte(page.Content), []byte(password))
	if err == nil {
		page.Content = string(c)
	}
	c, err = e.encrypt([]byte(page.Summary), []byte(password))
	if err == nil {
		page.Summary = string(c)
	}
	return page
}

func (e *Encrypt) Name() string {
	return "encrypt"
}

func newEncrypy(conf config.Config, theme theme.Theme) hook.Hook {
	return &Encrypt{conf: conf}
}
