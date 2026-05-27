package platform

import (
	"encoding/json"

	"github.com/lihaiya/aiops/pkg/aiops/password"
	"github.com/lihaiya/aiops/utils"
)

// checkBcrypt 使用 bcrypt 比对密码。
func checkBcrypt(pwd, hash string) bool {
	return utils.BcryptCheck(pwd, hash)
}

// hashBcrypt 使用 bcrypt 哈希密码。
func hashBcrypt(pwd string) string {
	return utils.BcryptHash(pwd)
}

// isPasswordInHistory 检查新密码是否在历史中（REQ-101）。
func isPasswordInHistory(newPassword, historyJSON string) bool {
	if historyJSON == "" {
		return false
	}
	var hashes []string
	if err := json.Unmarshal([]byte(historyJSON), &hashes); err != nil {
		return false
	}
	for _, hash := range hashes {
		if checkBcrypt(newPassword, hash) {
			return true
		}
	}
	return false
}

// appendPasswordHistory 追加密码到历史（REQ-101）。
func appendPasswordHistory(historyJSON, currentHash string, maxCount int) string {
	var hashes []string
	if historyJSON != "" {
		_ = json.Unmarshal([]byte(historyJSON), &hashes)
	}
	hashes = append(hashes, currentHash)
	if len(hashes) > maxCount {
		hashes = hashes[len(hashes)-maxCount:]
	}
	data, _ := json.Marshal(hashes)
	return string(data)
}

// getPasswordPolicy 获取密码策略（REQ-101）。
func getPasswordPolicy() *password.Policy {
	p := password.GetPolicy()
	return &p
}
