package api

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
)

// ---- 密码哈希:Argon2id(OWASP 推荐,内存困难型抗 GPU 破解) ----
// 兼容旧 bcrypt 哈希:verifyPassword 自动识别 bcrypt 前缀用 bcrypt 验证,
// 验证通过后调用方可选 upgradePassword 平滑迁移为 Argon2id。

const (
	argonTime    = 1            // 迭代次数
	argonMemory  = 64 * 1024    // 64 MiB 内存
	argonThreads = 4            // 并行度
	argonKeyLen  = 32           // 输出密钥长度
	argonSaltLen = 16           // 盐长度
)

// hashPassword 用 Argon2id 生成密码哈希,格式: $argon2id$v=19$m=65536,t=1,p=4$<salt>$<hash>
func hashPassword(password string) (string, error) {
	salt := make([]byte, argonSaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	key := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, argonMemory, argonTime, argonThreads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key)), nil
}

// verifyPassword 验证密码,自动识别 Argon2id / bcrypt 两种哈希格式。
func verifyPassword(hash, password string) bool {
	if strings.HasPrefix(hash, "$argon2id$") {
		return verifyArgon2(hash, password)
	}
	// 旧 bcrypt 哈希兼容
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// isArgon2Hash 判断哈希是否为 Argon2id 格式(用于登录后是否需要升级)。
func isArgon2Hash(hash string) bool {
	return strings.HasPrefix(hash, "$argon2id$")
}

// verifyArgon2 解析并验证 Argon2id 哈希。
func verifyArgon2(hash, password string) bool {
	parts := strings.Split(hash, "$")
	// ["", "argon2id", "v=19", "m=65536,t=1,p=4", salt, key]
	if len(parts) != 6 || !strings.HasPrefix(parts[1], "argon2id") {
		return false
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false
	}
	expected, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false
	}
	var params struct {
		m, t, p uint32
	}
	// 解析 m=...,t=...,p=...
	for _, kv := range strings.Split(parts[3], ",") {
		pair := strings.SplitN(kv, "=", 2)
		if len(pair) != 2 {
			return false
		}
		switch pair[0] {
		case "m":
			_, _ = fmt.Sscanf(pair[1], "%d", &params.m)
		case "t":
			_, _ = fmt.Sscanf(pair[1], "%d", &params.t)
		case "p":
			_, _ = fmt.Sscanf(pair[1], "%d", &params.p)
		}
	}
	if params.m == 0 || params.t == 0 || params.p == 0 {
		return false
	}
	computed := argon2.IDKey([]byte(password), salt, params.t, params.m, uint8(params.p), uint32(len(expected)))
	return subtle.ConstantTimeCompare(computed, expected) == 1
}
