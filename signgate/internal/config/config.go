package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	AppEnv              string
	HTTPPort            string
	PostgresURL         string
	RedisAddr           string
	ChainID             int64
	RPCURL              string
	AdapterAddress      string
	HookAddress         string
	EntrypointAddress   string
	RailguardSignerKey  string
	ReceiptSignerKey    string
	SignerKeyID         string
	OPAPolicyPath       string
	APIKey              string
	AllowNoopStore      bool
	WatcherEnabled      bool
	WatcherConfirmation int64
	WatcherPollSeconds  int64
	WatcherStartBlock   int64
	WatcherRescanBlocks int64
	UserOpStaleSeconds  int64
	BudgetAuthority     string
	KMSSignerEnabled    bool
	KMSSignerKeyID      string
}

func Load() Config {
	chainID, _ := strconv.ParseInt(getEnv("CHAIN_ID", "84532"), 10, 64)
	confirm, _ := strconv.ParseInt(getEnv("WATCHER_CONFIRMATION_DEPTH", "1"), 10, 64)
	poll, _ := strconv.ParseInt(getEnv("WATCHER_POLL_SECONDS", "5"), 10, 64)
	startBlock, _ := strconv.ParseInt(getEnv("WATCHER_START_BLOCK", "0"), 10, 64)
	rescan, _ := strconv.ParseInt(getEnv("WATCHER_RESCAN_BLOCKS", "12"), 10, 64)
	stale, _ := strconv.ParseInt(getEnv("USEROP_STALE_SECONDS", "300"), 10, 64)
	watcherEnabled := getEnv("WATCHER_ENABLED", "true") == "true"
	appEnv := getEnv("APP_ENV", "local")
	apiKey := getEnv("SIGNGATE_API_KEY", "")
	if apiKey == "" && strings.EqualFold(appEnv, "local") {
		apiKey = DevAPIKey
	}

	return Config{
		AppEnv:              appEnv,
		HTTPPort:            getEnv("HTTP_PORT", "8080"),
		PostgresURL:         getEnv("POSTGRES_URL", "postgres://railguard:railguard@localhost:5432/railguard?sslmode=disable"),
		RedisAddr:           getEnv("REDIS_ADDR", "localhost:6379"),
		ChainID:             chainID,
		RPCURL:              getEnv("RPC_URL", "http://localhost:8545"),
		AdapterAddress:      getEnv("ADAPTER_ADDRESS", ""),
		HookAddress:         getEnv("HOOK_ADDRESS", ""),
		EntrypointAddress:   getEnv("ENTRYPOINT_ADDRESS", ""),
		RailguardSignerKey:  getEnv("RAILGUARD_SIGNER_PRIVATE_KEY", ""),
		ReceiptSignerKey:    getEnv("RECEIPT_SIGNER_PRIVATE_KEY", ""),
		SignerKeyID:         getEnv("SIGNER_KEY_ID", "railguard-key-v1"),
		OPAPolicyPath:       getEnv("OPA_POLICY_PATH", "../policy/railguard.rego"),
		APIKey:              apiKey,
		AllowNoopStore:      getEnv("ALLOW_NOOP_STORE", "false") == "true",
		WatcherEnabled:      watcherEnabled,
		WatcherConfirmation: confirm,
		WatcherPollSeconds:  poll,
		WatcherStartBlock:   startBlock,
		WatcherRescanBlocks: rescan,
		UserOpStaleSeconds:  stale,
		BudgetAuthority:     getEnv("BUDGET_AUTHORITY", "redis"),
		KMSSignerEnabled:    getEnv("KMS_SIGNER_ENABLED", "false") == "true",
		KMSSignerKeyID:      getEnv("KMS_SIGNER_KEY_ID", ""),
	}
}

func (c Config) IsLocal() bool {
	return strings.EqualFold(c.AppEnv, "local")
}

func (c Config) UsePostgresBudgetAuthority() bool {
	return strings.EqualFold(c.BudgetAuthority, "postgres")
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
