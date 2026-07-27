package encrypt

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/honmaple/snow/internal/core"
	"github.com/honmaple/snow/internal/site/content"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func decryptForTest(t *testing.T, enc *Encrypt, ciphertext string, password string) string {
	t.Helper()

	ekey := enc.deriveKey(password)
	block, err := aes.NewCipher(ekey)
	require.NoError(t, err)

	data, err := base64.StdEncoding.DecodeString(ciphertext)
	require.NoError(t, err)

	iv := enc.ivForBlockSize(block.BlockSize(), ekey)
	switch enc.mode {
	case modeCBC:
		plain := make([]byte, len(data))
		cipher.NewCBCDecrypter(block, iv).CryptBlocks(plain, data)
		padding := int(plain[len(plain)-1])
		return string(plain[:len(plain)-padding])
	case modeCTR:
		plain := make([]byte, len(data))
		cipher.NewCTR(block, iv).XORKeyStream(plain, data)
		return string(plain)
	case modeCFB:
		plain := make([]byte, len(data))
		cipher.NewCFBDecrypter(block, iv).XORKeyStream(plain, data)
		return string(plain)
	case modeOFB:
		plain := make([]byte, len(data))
		cipher.NewOFB(block, iv).XORKeyStream(plain, data)
		return string(plain)
	default:
		t.Fatalf("unsupported mode %s", enc.mode)
		return ""
	}
}

func TestNewEncryptDefaultsToCBC(t *testing.T) {
	enc, err := NewEncrypt("")
	require.NoError(t, err)
	assert.Equal(t, modeCBC, enc.mode)
}

func TestNewEncryptRejectsUnknownMode(t *testing.T) {
	_, err := NewEncrypt("ecb")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `unsupported encrypt mode "ecb"`)
}

func TestEncryptSupportsAESModes(t *testing.T) {
	password := "secret"
	plaintext := "hello snow"
	modes := []string{modeCBC, modeCTR, modeCFB, modeOFB}

	results := make(map[string]string, len(modes))
	for _, mode := range modes {
		t.Run(mode, func(t *testing.T) {
			enc, err := NewEncrypt(mode)
			require.NoError(t, err)

			ciphertext, err := enc.Encrypt(plaintext, password)
			require.NoError(t, err)
			assert.NotEmpty(t, ciphertext)
			assert.NotEqual(t, plaintext, ciphertext)
			assert.Equal(t, plaintext, decryptForTest(t, enc, ciphertext, password))

			results[mode] = ciphertext
		})
	}

	assert.NotEqual(t, results[modeCBC], results[modeCTR])
	assert.NotEqual(t, results[modeCBC], results[modeCFB])
	assert.NotEqual(t, results[modeCBC], results[modeOFB])
}

func TestEncryptUsesCustomIV(t *testing.T) {
	enc, err := NewEncrypt(modeCTR)
	require.NoError(t, err)
	enc.iv = bytes.Repeat([]byte{1}, aes.BlockSize)

	ciphertext, err := enc.Encrypt("hello", "secret")
	require.NoError(t, err)

	assert.Equal(t, "hello", decryptForTest(t, enc, ciphertext, "secret"))
}

func TestEncryptHookAddsModeToWrappedShortcode(t *testing.T) {
	h := &EncryptHook{
		opt: Option{
			Mode:        modeCFB,
			Description: "desc",
		},
	}
	page := &content.Page{
		Node: &content.Node{
			FrontMatter: content.NewFrontMatter(map[string]any{"password": "123"}),
			Content:     "body",
			Summary:     "summary",
		},
	}

	page = h.HandlePage(page)

	assert.Contains(t, page.Content, `mode="cfb"`)
	assert.Contains(t, page.Summary, `mode="cfb"`)
}

func TestNewValidatesConfiguredMode(t *testing.T) {
	conf := core.DefaultConfig()
	conf.Set("hooks.encrypt.option.mode", "ctr")
	ctx, err := core.NewContext(conf)
	require.NoError(t, err)

	h, err := New(ctx)
	require.NoError(t, err)
	assert.Equal(t, modeCTR, h.(*EncryptHook).opt.Mode)
}

func TestNewRejectsInvalidConfiguredMode(t *testing.T) {
	conf := core.DefaultConfig()
	conf.Set("hooks.encrypt.option.mode", "ecb")
	ctx, err := core.NewContext(conf)
	require.NoError(t, err)

	_, err = New(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `unsupported encrypt mode "ecb"`)
}

func TestDeriveKeyRemainsStable(t *testing.T) {
	enc, err := NewEncrypt(modeCBC)
	require.NoError(t, err)

	assert.Equal(t, []byte(fmt.Sprintf("%x", md5.Sum([]byte("secret")))), enc.deriveKey("secret"))
}
