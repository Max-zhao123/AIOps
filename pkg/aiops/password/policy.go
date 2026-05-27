package password

import (
	"errors"
	"fmt"
	"unicode"

	"go.uber.org/zap"
)

// Policy 密码策略配置（REQ-101）。
type Policy struct {
	MinLength            int  `json:"minLength" yaml:"minLength"`
	RequireUppercase    bool `json:"requireUppercase" yaml:"requireUppercase"`
	RequireLowercase    bool `json:"requireLowercase" yaml:"requireLowercase"`
	RequireDigit        bool `json:"requireDigit" yaml:"requireDigit"`
	RequireSpecialChar  bool `json:"requireSpecialChar" yaml:"requireSpecialChar"`
	MaxPasswordAgeDays  int  `json:"maxPasswordAgeDays" yaml:"maxPasswordAgeDays"`
	PasswordHistoryCount int `json:"passwordHistoryCount" yaml:"passwordHistoryCount"`
	MaxFailedLogins      int  `json:"maxFailedLogins" yaml:"maxFailedLogins"`
	LockoutDurationMin   int  `json:"lockoutDurationMin" yaml:"lockoutDurationMin"`
}

// DefaultPolicy 默认密码策略。
var DefaultPolicy = Policy{
	MinLength:            12,
	RequireUppercase:     true,
	RequireLowercase:     true,
	RequireDigit:         true,
	RequireSpecialChar:   true,
	MaxPasswordAgeDays:   90,
	PasswordHistoryCount: 5,
	MaxFailedLogins:       5,
	LockoutDurationMin:   30,
}

// Validate 验证密码是否满足策略要求。
func (p Policy) Validate(password string) error {
	if len(password) < p.MinLength {
		return fmt.Errorf("密码不满足复杂度要求：至少%d位", p.MinLength)
	}
	if len(password) > 128 {
		return errors.New("密码最大128字节")
	}
	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSpecial = true
		}
	}
	var missing []string
	if p.RequireUppercase && !hasUpper {
		missing = append(missing, "大写字母")
	}
	if p.RequireLowercase && !hasLower {
		missing = append(missing, "小写字母")
	}
	if p.RequireDigit && !hasDigit {
		missing = append(missing, "数字")
	}
	if p.RequireSpecialChar && !hasSpecial {
		missing = append(missing, "特殊字符")
	}
	if len(missing) > 0 {
		return fmt.Errorf("密码不满足复杂度要求：至少%d位，含%s", p.MinLength, joinChinese(missing))
	}
	return nil
}

// joinChinese 用顿号连接中文字符串。
func joinChinese(items []string) string {
	result := ""
	for i, item := range items {
		if i > 0 {
			result += "、"
		}
		result += item
	}
	return result
}

// GetPolicy 从 viper 配置读取密码策略，未配置则使用默认值。
func GetPolicy() Policy {
	p := DefaultPolicy
	// 尝试从 viper 读取，这里简化为返回默认值
	// 实际部署时可从 config.yaml 读取
	return p
}

// LogPasswordValidation 记录密码验证失败日志（不记录密码原文）。
func LogPasswordValidation(err error) {
	if err != nil {
		zap.L().Warn("密码验证失败", zap.String("reason", err.Error()))
	}
}
