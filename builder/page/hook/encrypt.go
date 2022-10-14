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

func (e *Encrypt) Name() string {
	return "encrypt"
}

func (e *Encrypt) Write(page *page.Page) *page.Page {
	password := page.Meta["password"]
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

func (e *Encrypt) pkcs7Padding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

func (e *Encrypt) pkcs7UnPadding(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, errors.New("err password")
	}
	unPadding := int(data[length-1])
	return data[:(length - unPadding)], nil
}

func (e *Encrypt) encrypt(data []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	encryptBytes := e.pkcs7Padding(data, blockSize)
	crypted := make([]byte, len(encryptBytes))
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	blockMode.CryptBlocks(crypted, encryptBytes)
	return crypted, nil
}

func (e *Encrypt) decrypt(data []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	crypted := make([]byte, len(data))
	blockMode.CryptBlocks(crypted, data)
	crypted, err = e.pkcs7UnPadding(crypted)
	if err != nil {
		return nil, err
	}
	return crypted, nil
}

func newEncrypy(conf config.Config, theme theme.Theme) hook.Hook {
	return &Encrypt{conf: conf}
}
