package crypto

import (
	"encoding/base64"
	"strings"
	"testing"
)

const testKey = "test-encryption-key"

func TestEncryptDecryptRoundtrip(t *testing.T) {
	cases := []struct {
		name      string
		plaintext string
		key       string
	}{
		{"普通 API Key", "sk-abc1234567890xyz", testKey},
		{"空明文", "", testKey},
		{"中文与符号", "密钥-🔑-!@#$%^&*()", testKey},
		{"超过 32 字节的密钥（截断）", "secret", strings.Repeat("k", 40)},
		{"长明文", strings.Repeat("a", 4096), testKey},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			encrypted, err := Encrypt(tc.plaintext, tc.key)
			if err != nil {
				t.Fatalf("Encrypt 失败: %v", err)
			}
			decrypted, err := Decrypt(encrypted, tc.key)
			if err != nil {
				t.Fatalf("Decrypt 失败: %v", err)
			}
			if decrypted != tc.plaintext {
				t.Errorf("往返结果不一致: got %q, want %q", decrypted, tc.plaintext)
			}
		})
	}
}

func TestDecryptWrongKey(t *testing.T) {
	encrypted, err := Encrypt("sk-abc1234567890xyz", testKey)
	if err != nil {
		t.Fatalf("Encrypt 失败: %v", err)
	}
	if _, err := Decrypt(encrypted, "another-key"); err == nil {
		t.Error("错误密钥解密应当失败，但成功了")
	}
}

func TestDecryptTamperedCiphertext(t *testing.T) {
	encrypted, err := Encrypt("sk-abc1234567890xyz", testKey)
	if err != nil {
		t.Fatalf("Encrypt 失败: %v", err)
	}
	raw, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		t.Fatalf("解码密文失败: %v", err)
	}
	raw[len(raw)-1] ^= 0xFF // 翻转最后一个字节
	tampered := base64.StdEncoding.EncodeToString(raw)
	if _, err := Decrypt(tampered, testKey); err == nil {
		t.Error("篡改后的密文解密应当失败（GCM 认证），但成功了")
	}
}

func TestDecryptInvalidInput(t *testing.T) {
	if _, err := Decrypt("not-base64!!!", testKey); err == nil {
		t.Error("非法 base64 应当返回错误")
	}
	// 合法 base64 但长度不足一个 nonce
	short := base64.StdEncoding.EncodeToString([]byte("abc"))
	if _, err := Decrypt(short, testKey); err == nil {
		t.Error("长度不足的密文应当返回错误")
	}
}

func TestMaskKey(t *testing.T) {
	if got := MaskKey("12345678abcd"); got != "12345678***" {
		t.Errorf("MaskKey 长 key: got %q", got)
	}
	if got := MaskKey("short"); got != "***" {
		t.Errorf("MaskKey 短 key: got %q", got)
	}
}

func TestMaskPlainKey(t *testing.T) {
	if got := MaskPlainKey("sk-abc1234567890xyz"); got != "sk-abc***0xyz" {
		t.Errorf("MaskPlainKey 长 key: got %q", got)
	}
	if got := MaskPlainKey("sk-short10"); got != "***" {
		t.Errorf("MaskPlainKey 短 key 应整体脱敏: got %q", got)
	}
}
