package api

import (
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestArgon2HashRoundTrip(t *testing.T) {
	hash, err := hashPassword("MySecret123!")
	if err != nil {
		t.Fatalf("hashPassword: %v", err)
	}
	if !strings.HasPrefix(hash, "$argon2id$") {
		t.Fatalf("expected argon2id prefix, got %s", hash[:20])
	}
	if !verifyPassword(hash, "MySecret123!") {
		t.Fatal("correct password should verify")
	}
	if verifyPassword(hash, "WrongPass!") {
		t.Fatal("wrong password should not verify")
	}
}

func TestVerifyBcryptCompat(t *testing.T) {
	// 模拟旧 bcrypt 哈希(数据库里已存在的用户)
	oldHash, err := bcrypt.GenerateFromPassword([]byte("legacy-pass-123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("bcrypt generate: %v", err)
	}
	if !verifyPassword(string(oldHash), "legacy-pass-123") {
		t.Fatal("bcrypt hash should verify with legacy password")
	}
	if verifyPassword(string(oldHash), "wrong") {
		t.Fatal("bcrypt hash should reject wrong password")
	}
	if isArgon2Hash(string(oldHash)) {
		t.Fatal("bcrypt hash should not be detected as argon2")
	}
}

func TestVerifyArgon2Malformed(t *testing.T) {
	if verifyPassword("$argon2id$bad", "x") {
		t.Fatal("malformed argon2 hash should not verify")
	}
	if verifyPassword("", "x") {
		t.Fatal("empty hash should not verify")
	}
}
