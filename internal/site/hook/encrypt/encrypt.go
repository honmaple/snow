package encrypt

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/site/content"
	"github.com/honmaple/snow/internal/site/hook"
	"github.com/honmaple/snow/internal/site/template"
)

type (
	Option struct {
		Password    string `json:"password"`
		Description string `json:"description"`
	}
	EncryptHook struct {
		hook.HookImpl
		ctx *core.Context
		opt Option
		iv  []byte
	}
)

func (e *EncryptHook) deriveKey(key string) []byte {
	has := md5.Sum([]byte(key))
	return []byte(fmt.Sprintf("%x", has))
}

func (e *EncryptHook) pkcs7Padding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

func (e *EncryptHook) encrypt(plaintext, key string) (string, error) {
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

func (e *EncryptHook) HandlePage(page *content.Page) *content.Page {
	password := page.FrontMatter.GetString("password")
	if password == "" {
		return page
	}
	description := e.opt.Description
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

func (e *EncryptHook) HandleTemplate(set template.TemplateSet) error {
	set.RegisterFilter("encrypt", e.encryptFilter)
	return nil
}

func New(ctx *core.Context) (hook.Hook, error) {
	var opt Option
	if err := hook.Unmarshal(ctx.Config.Get("hooks.encrypt.option"), &opt); err != nil {
		return nil, err
	}

	e := &EncryptHook{
		ctx: ctx,
		opt: opt,
	}
	return e, nil
}

func init() {
	hook.Register("encrypt", New)
}
