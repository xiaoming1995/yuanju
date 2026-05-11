package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

// Encrypt 使用 AES-256-GCM 加密明文，返回 base64 编码的密文
func Encrypt(plaintext, keyStr string) (string, error) {
	key := padKey(keyStr)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt 解密 base64 编码的密文，返回明文
func Decrypt(encoded, keyStr string) (string, error) {
	key := padKey(keyStr)
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(data) < gcm.NonceSize() {
		return "", errors.New("密文长度不足")
	}
	nonce, ciphertext := data[:gcm.NonceSize()], data[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

// MaskKey 将 API Key 脱敏展示（前8位 + ***)
func MaskKey(key string) string {
	if len(key) <= 8 {
		return "***"
	}
	return key[:8] + "***"
}

// MaskPlainKey 将明文 API Key 脱敏：前6位 + *** + 后4位
// 例：sk-abc1234567890xyz → sk-abc***0xyz
func MaskPlainKey(plaintext string) string {
	if len(plaintext) <= 10 {
		return "***"
	}
	return plaintext[:6] + "***" + plaintext[len(plaintext)-4:]
}

// padKey 将密钥补齐/截断到 32 字节
func padKey(key string) []byte {
	b := []byte(key)
	if len(b) >= 32 {
		return b[:32]
	}
	padded := make([]byte, 32)
	copy(padded, b)
	return padded
}
