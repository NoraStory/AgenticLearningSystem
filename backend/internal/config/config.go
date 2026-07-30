package config

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Config struct {
	Port                 string
	DatabaseURL          string
	RedisAddr            string
	RedisPassword        string
	MinIOEndpoint        string
	MinIOAccessKey       string
	MinIOSecretKey       string
	MinIOUseSSL          bool
	MinIOBucket          string
	JWTSecret            string
	CORSOrigins          []string
	ArkAPIKey            string
	ArkBaseURL           string
	LLMWireAPI           string
	ArkModel             string
	UploadDir            string
	SearchProvider       string
	SearchBaseURL        string
	SearchTimeoutSeconds int
}

func Load() Config {
	loadEnvFiles(".env", filepath.Join("..", ".env"))
	return Config{
		Port:                 getenv("PORT", "8080"),
		DatabaseURL:          getenv("DATABASE_URL", "postgres://codeforge:codeforge@localhost:5432/codeforge?sslmode=disable"),
		RedisAddr:            getenv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:        os.Getenv("REDIS_PASSWORD"),
		MinIOEndpoint:        getenv("MINIO_ENDPOINT", "localhost:9000"),
		MinIOAccessKey:       getenv("MINIO_ACCESS_KEY", "codeforge"),
		MinIOSecretKey:       getenv("MINIO_SECRET_KEY", "codeforge123"),
		MinIOUseSSL:          getbool("MINIO_USE_SSL", false),
		MinIOBucket:          getenv("MINIO_BUCKET", "codeforge"),
		JWTSecret:            getenv("JWT_SECRET", "change-me-in-production-codeforge-secret"),
		CORSOrigins:          strings.Split(getenv("CORS_ORIGINS", "http://localhost:5000,http://127.0.0.1:5000"), ","),
		ArkAPIKey:            getenv("ARK_API_KEY", os.Getenv("OPENAI_API_KEY")),
		ArkBaseURL:           getenv("ARK_BASE_URL", "https://ark.cn-beijing.volces.com/api/v3"),
		ArkModel:             getenv("ARK_MODEL", getenv("MODEL", "")),
		LLMWireAPI:           getenv("LLM_WIRE_API", "chat_completions"),
		UploadDir:            getenv("UPLOAD_DIR", "./data/uploads"),
		SearchProvider:       getenv("SEARCH_PROVIDER", "searxng"),
		SearchBaseURL:        getenv("SEARCH_BASE_URL", "http://localhost:8081"),
		SearchTimeoutSeconds: getint("SEARCH_TIMEOUT_SECONDS", 8),
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
func getint(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 1 {
		return fallback
	}
	return n
}

func getbool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	b, e := strconv.ParseBool(v)
	if e != nil {
		return fallback
	}
	return b
}

func loadEnvFiles(paths ...string) {
	for _, path := range paths {
		file, err := os.Open(path)
		if err != nil {
			continue
		}
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 && os.Getenv(strings.TrimSpace(parts[0])) == "" {
				_ = os.Setenv(strings.TrimSpace(parts[0]), strings.Trim(strings.TrimSpace(parts[1]), `"'`))
			}
		}
		_ = file.Close()
	}
}
