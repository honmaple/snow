package encrypt

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	// "crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/flosch/pongo2/v6"
	// "golang.org/x/crypto/pbkdf2"

	"crypto/md5"
	"github.com/honmaple/snow/builder/hook"
	"github.com/honmaple/snow/builder/page"
	"github.com/honmaple/snow/builder/theme"
	"github.com/honmaple/snow/config"
)

type Encrypt struct {
	hook.BaseHook
	conf config.Config
	iv   []byte
}

func (e *Encrypt) deriveKey(key string) []byte {
	// return pbkdf2.Key([]byte(key), e.salt, 1000, 32, sha256.New)
	has := md5.Sum([]byte(key))
	return []byte(fmt.Sprintf("%x", has))

	// h := md5.New()
	// h.Write([]byte(key))
	// return h.Sum(nil)
	// return hex.En(h.Sum(nil))
}

func (e *Encrypt) pkcs7Padding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

func (e *Encrypt) pkcs7UnPadding(data []byte) []byte {
	length := len(data)
	return data[:(length - int(data[length-1]))]
}

func (e *Encrypt) encrypt(plaintext, key string) (string, error) {
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

func (e *Encrypt) decrypt(ciphertext, key string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}
	dkey := e.deriveKey(key)

	block, err := aes.NewCipher(dkey)
	if err != nil {
		return "", err
	}
	blockSize := block.BlockSize()
	iv := e.iv
	if len(iv) != blockSize {
		iv = dkey[:blockSize]
	}
	blockMode := cipher.NewCBCDecrypter(block, iv)

	decrypted := make([]byte, len(data))
	blockMode.CryptBlocks(decrypted, data)
	decrypted = e.pkcs7UnPadding(decrypted)
	return string(decrypted), nil
}

func (e *Encrypt) AfterPageParse(page *page.Page) *page.Page {
	password := page.Get("password")
	if password == nil || password == "" {
		return page
	}
	page.Summary = fmt.Sprintf(`<shortcode _name="encrypt" password="%s">%s</shortcode>`, password, page.Summary)
	page.Content = fmt.Sprintf(`<shortcode _name="encrypt" password="%s">%s</shortcode>`, password, page.Content)
	return page
}

func (e *Encrypt) Name() string {
	return "encrypt"
}

func (e *Encrypt) filter(in *pongo2.Value, param *pongo2.Value) (out *pongo2.Value, err *pongo2.Error) {
	plaintext, ok := in.Interface().(string)
	if !ok {
		return nil, &pongo2.Error{
			Sender:    "filter:encrypt",
			OrigError: errors.New("filter input argument must be of type 'string'"),
		}
	}
	if param == nil {
		return nil, &pongo2.Error{
			Sender:    "filter:encrypt",
			OrigError: errors.New("password is required"),
		}
	}
	password := param.String()

	text, err1 := e.encrypt(plaintext, password)
	if err1 != nil {
		return nil, &pongo2.Error{
			Sender:    "filter:encrypt",
			OrigError: err1,
		}
	}
	return pongo2.AsValue(text), nil
}

func New(conf config.Config, theme theme.Theme) hook.Hook {
	e := &Encrypt{
		conf: conf,
	}

	pongo2.RegisterFilter("encrypt", e.filter)
	return e
}

func init() {
	hook.Register("encrypt", New)
}
