package encrypt

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/honmaple/snow/internal/site/content"
	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/site/hook"
)

type Encrypt struct {
	hook.HookImpl
	ctx *core.Context
	iv  []byte
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

func (e *Encrypt) HandlePage(page *content.Page) *content.Page {
	password := page.FrontMatter.GetString("password")
	if password == "" {
		return page
	}
	description := e.ctx.Config.GetString("hooks.encrypt.desc")
	if v := strings.SplitN(password, ",", 2); len(v) == 2 {
		password = v[0]
		description = v[1]
	}
	if description == "" {
		description = "这是一篇加密的文章，你需要输入正确的密码."
	}
	page.Summary = fmt.Sprintf(`<shortcode _name="encrypt" password="%s" description="%s">%s</shortcode>`, password, description, page.Summary)
	page.Content = fmt.Sprintf(`<shortcode _name="encrypt" password="%s" description="%s">%s</shortcode>`, password, description, page.Content)
	return page
}

func New(ctx *core.Context) (hook.Hook, error) {
	e := &Encrypt{
		ctx: ctx,
	}
	return e, nil
}

func init() {
	hook.Register("encrypt", New)
}
