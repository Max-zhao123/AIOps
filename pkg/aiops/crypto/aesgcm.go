package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"sync"

	"go.uber.org/zap"
)

var (
	// encryptionKey 从环境变量 AIOPS_ENCRYPTION_KEY 加载（REQ-093）。
	encryptionKey []byte

	// keyMu 保护 encryptionKey 的读写操作，确保密钥切换的原子性。
	keyMu sync.Mutex
)

// InitKey 从环境变量加载加密密钥。必须在服务启动时调用。
// 密钥必须为 32 字节（AES-256）的 base64 编码字符串。
func InitKey() error {
	keyStr := os.Getenv("AIOPS_ENCRYPTION_KEY")
	if keyStr == "" {
		return errors.New("环境变量 AIOPS_ENCRYPTION_KEY 未设置，服务启动失败")
	}
	decoded, err := base64.StdEncoding.DecodeString(keyStr)
	if err != nil {
		return fmt.Errorf("AIOPS_ENCRYPTION_KEY base64 解码失败: %w", err)
	}
	if len(decoded) != 32 {
		return fmt.Errorf("AIOPS_ENCRYPTION_KEY 解码后长度必须为 32 字节，实际 %d 字节", len(decoded))
	}
	keyMu.Lock()
	encryptionKey = decoded
	keyMu.Unlock()
	return nil
}

// Encrypt AES-256-GCM 加密明文，返回 base64 编码密文。
func Encrypt(plaintext string) (string, error) {
	if len(plaintext) > 4096 {
		return "", errors.New("凭证值最大 4096 字节")
	}
	keyMu.Lock()
	key := encryptionKey
	keyMu.Unlock()
	if len(key) == 0 {
		return "", errors.New("加密密钥未初始化")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt AES-256-GCM 解密 base64 编码密文，返回明文。
func Decrypt(encoded string) (string, error) {
	keyMu.Lock()
	key := encryptionKey
	keyMu.Unlock()
	if len(key) == 0 {
		return "", errors.New("加密密钥未初始化")
	}
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("base64 解码失败: %w", err)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("密文长度不足")
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

// ReEncrypt 使用新密钥重新加密所有凭证（密钥轮转）。
// newKeyBase64 是新密钥的 base64 编码。
// 使用 sync.Mutex 保护密钥切换的原子性，避免并发调用 Encrypt/Decrypt 时使用错误密钥。
func ReEncrypt(oldEncrypted string, newKeyBase64 string) (string, error) {
	// 先用旧密钥解密（在锁内读取旧密钥并完成解密）
	keyMu.Lock()
	oldKey := make([]byte, len(encryptionKey))
	copy(oldKey, encryptionKey)
	keyMu.Unlock()

	// 使用旧密钥解密：手动构造 cipher，不依赖全局变量
	data, err := base64.StdEncoding.DecodeString(oldEncrypted)
	if err != nil {
		return "", fmt.Errorf("base64 解码失败: %w", err)
	}
	block, err := aes.NewCipher(oldKey)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("密文长度不足")
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	// 解码新密钥
	newKey, err := base64.StdEncoding.DecodeString(newKeyBase64)
	if err != nil {
		return "", fmt.Errorf("新密钥 base64 解码失败: %w", err)
	}
	if len(newKey) != 32 {
		return "", fmt.Errorf("新密钥长度必须为 32 字节")
	}

	// 使用新密钥加密：直接构造 cipher，不依赖全局变量
	newBlock, err := aes.NewCipher(newKey)
	if err != nil {
		return "", err
	}
	newGcm, err := cipher.NewGCM(newBlock)
	if err != nil {
		return "", err
	}
	newNonce := make([]byte, newGcm.NonceSize())
	if _, err := rand.Read(newNonce); err != nil {
		return "", err
	}
	reEncrypted := newGcm.Seal(newNonce, newNonce, plaintext, nil)

	return base64.StdEncoding.EncodeToString(reEncrypted), nil
}

// GetKey 返回当前加密密钥（仅用于 re-encrypt 流程）。
func GetKey() []byte {
	keyMu.Lock()
	defer keyMu.Unlock()
	return encryptionKey
}

// SetKey 设置加密密钥（仅用于 re-encrypt 流程和测试）。
func SetKey(key []byte) {
	keyMu.Lock()
	encryptionKey = key
	keyMu.Unlock()
}

// MustInitKey 初始化密钥，失败则 panic（用于 platform main）。
func MustInitKey() {
	if err := InitKey(); err != nil {
		zap.L().Error("加密密钥初始化失败", zap.Error(err))
		panic(err)
	}
}
