package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Config aggregates all configuration for the application.
type Config struct {
	App       AppConfig
	Database  DatabaseConfig
	Auth      AuthConfig
	Turnstile TurnstileConfig
	Redis     RedisConfig
	GeoIP     GeoIPConfig
	Backup    BackupConfig
	Export    ExportConfig
	Media     MediaConfig
}

// AppConfig contains Fiber specific settings.
type AppConfig struct {
	Name                     string
	Port                     string
	Env                      string
	HTMLSnapshotBaseURL      string
	ProxyHeader              string
	TrustedProxies           []string
	TrustedProxyCheck        bool
	IPValidation             bool
	UpdateCheckEnabled       bool
	UpdateCheckRepo          string
	UpdateCheckChannel       string
	TelemetryDefaultEndpoint string
}

// DatabaseConfig captures everything required to boot GORM.
type DatabaseConfig struct {
	Driver      string
	DSN         string
	AutoMigrate bool
}

// AuthConfig 控制 JWT 签发与校验。
type AuthConfig struct {
	Secret        string
	Issuer        string
	AccessTTL     time.Duration
	OAuthStateTTL time.Duration
}

// TurnstileConfig 控制 Cloudflare Turnstile 人机校验。
type TurnstileConfig struct {
	Enabled   bool
	Secret    string
	VerifyURL string
	Timeout   time.Duration
}

// RedisConfig 描述 Redis 连接。
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
	Prefix   string
}

// GeoIPConfig 描述 IP 归属地数据库配置。
type GeoIPConfig struct {
	DBPath      string
	DownloadURL string
	ASNPath     string
	ASNURL      string
}

// BackupConfig controls whole-site archive creation and download tickets.
type BackupConfig struct {
	RootDir                  string
	UploadDir                string
	PGDumpBin                string
	PGRestoreBin             string
	TicketTTL                time.Duration
	CommandTimeout           time.Duration
	SchedulerPollInterval    time.Duration
	RestoreMaxArchiveBytes   int64
	RestoreMaxExtractedBytes int64
}

// ExportConfig controls the built-in content export (markdown + bundled images archive).
type ExportConfig struct {
	RootDir          string
	TicketTTL        time.Duration
	JobTimeout       time.Duration
	ExternalWorkers  int
	ExternalTimeout  time.Duration
	MaxExternalBytes int64
}

// MediaConfig controls optional R2 mirroring and signed delivery. Local files
// remain the durable fallback when R2 is disabled or temporarily unavailable.
type MediaConfig struct {
	R2Endpoint        string
	R2Bucket          string
	R2AccessKeyID     string
	R2SecretAccessKey string
	R2Prefix          string
	R2ReadURLTTL      time.Duration
	R2RequestTimeout  time.Duration
	R2UploadTimeout   time.Duration
}

// Load builds a Config struct with sane defaults overridden by environment variables.
func Load() Config {
	return Config{
		App: AppConfig{
			Name:                getEnv("APP_NAME", "shawn-blog-server"),
			Port:                getEnv("APP_PORT", "8080"),
			Env:                 strings.ToLower(getEnv("APP_ENV", "development")),
			HTMLSnapshotBaseURL: strings.TrimRight(getEnv("HTMLSNAPSHOT_BASE_URL", "http://localhost:3000"), "/"),
			ProxyHeader:         getEnv("APP_PROXY_HEADER", "X-Forwarded-For"),
			TrustedProxies: getEnvAsSlice("APP_TRUSTED_PROXIES", []string{
				"127.0.0.1",
				"::1",
				"10.0.0.0/8",
				"172.16.0.0/12",
				"192.168.0.0/16",
				"fc00::/7",
			}),
			TrustedProxyCheck:        getEnvAsBool("APP_TRUSTED_PROXY_CHECK", true),
			IPValidation:             getEnvAsBool("APP_IP_VALIDATION", true),
			UpdateCheckEnabled:       getEnvAsBool("APP_UPDATE_CHECK_ENABLED", false),
			UpdateCheckRepo:          strings.TrimSpace(getEnv("APP_UPDATE_CHECK_REPO", "shawns-yao/shawn-blog")),
			UpdateCheckChannel:       strings.TrimSpace(getEnv("APP_UPDATE_CHANNEL", "stable")),
			TelemetryDefaultEndpoint: strings.TrimSpace(getEnv("TELEMETRY_DEFAULT_ENDPOINT", "")),
		},
		Database: DatabaseConfig{
			Driver:      strings.ToLower(getEnv("DB_DRIVER", "postgres")),
			DSN:         getEnv("DB_DSN", "postgres://postgres:postgres@localhost:5432/shawn-blog?sslmode=disable"),
			AutoMigrate: getEnvAsBool("DB_AUTO_MIGRATE", true),
		},
		Auth: AuthConfig{
			Secret:        getEnv("AUTH_SECRET", "change-me"),
			Issuer:        getEnv("AUTH_ISSUER", "shawn-blog-api"),
			AccessTTL:     getEnvAsDuration("AUTH_ACCESS_TTL", 7*24*time.Hour),
			OAuthStateTTL: getEnvAsDuration("AUTH_STATE_TTL", time.Minute*10),
		},
		Turnstile: TurnstileConfig{
			Enabled:   getEnvAsBool("TURNSTILE_ENABLED", false),
			Secret:    getEnv("TURNSTILE_SECRET", ""),
			VerifyURL: getEnv("TURNSTILE_VERIFY_URL", "https://challenges.cloudflare.com/turnstile/v0/siteverify"),
			Timeout:   getEnvAsDuration("TURNSTILE_TIMEOUT", 5*time.Second),
		},
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "127.0.0.1:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
			Prefix:   getEnv("REDIS_PREFIX", "shawn-blog:"),
		},
		GeoIP: GeoIPConfig{
			DBPath:      getEnv("GEOIP_DB_PATH", "storage/geoip/GeoLite2-City.mmdb"),
			DownloadURL: getEnv("GEOIP_DB_URL", "https://github.com/P3TERX/GeoLite.mmdb/raw/download/GeoLite2-City.mmdb"),
			ASNPath:     getEnv("GEOIP_ASN_DB_PATH", "storage/geoip/GeoLite2-ASN.mmdb"),
			ASNURL:      getEnv("GEOIP_ASN_DB_URL", "https://github.com/P3TERX/GeoLite.mmdb/raw/download/GeoLite2-ASN.mmdb"),
		},
		Backup: BackupConfig{
			RootDir:                  getEnv("BACKUP_ROOT_DIR", "storage/backups"),
			UploadDir:                getEnv("BACKUP_UPLOAD_DIR", "storage/uploads"),
			PGDumpBin:                getEnv("BACKUP_PG_DUMP_BIN", "pg_dump"),
			PGRestoreBin:             getEnv("BACKUP_PG_RESTORE_BIN", "pg_restore"),
			TicketTTL:                getEnvAsDuration("BACKUP_DOWNLOAD_TICKET_TTL", 10*time.Minute),
			CommandTimeout:           getEnvAsDuration("BACKUP_COMMAND_TIMEOUT", 30*time.Minute),
			SchedulerPollInterval:    getEnvAsDuration("BACKUP_SCHEDULER_POLL_INTERVAL", 30*time.Second),
			RestoreMaxArchiveBytes:   getEnvAsInt64("BACKUP_RESTORE_MAX_ARCHIVE_BYTES", 10<<30),
			RestoreMaxExtractedBytes: getEnvAsInt64("BACKUP_RESTORE_MAX_EXTRACTED_BYTES", 50<<30),
		},
		Export: ExportConfig{
			RootDir:          getEnv("EXPORT_ROOT_DIR", "storage/exports"),
			TicketTTL:        getEnvAsDuration("EXPORT_DOWNLOAD_TICKET_TTL", 10*time.Minute),
			JobTimeout:       getEnvAsDuration("EXPORT_JOB_TIMEOUT", 30*time.Minute),
			ExternalWorkers:  int(getEnvAsInt64("EXPORT_EXTERNAL_WORKERS", 4)),
			ExternalTimeout:  getEnvAsDuration("EXPORT_EXTERNAL_TIMEOUT", 15*time.Second),
			MaxExternalBytes: getEnvAsInt64("EXPORT_MAX_EXTERNAL_BYTES", 25<<20),
		},
		Media: MediaConfig{
			R2Endpoint:        strings.TrimRight(strings.TrimSpace(getEnv("MEDIA_R2_ENDPOINT", "")), "/"),
			R2Bucket:          strings.TrimSpace(getEnv("MEDIA_R2_BUCKET", "")),
			R2AccessKeyID:     strings.TrimSpace(getEnv("MEDIA_R2_ACCESS_KEY_ID", "")),
			R2SecretAccessKey: strings.TrimSpace(getEnv("MEDIA_R2_SECRET_ACCESS_KEY", "")),
			R2Prefix:          strings.Trim(strings.TrimSpace(getEnv("MEDIA_R2_PREFIX", "blog/")), "/"),
			R2ReadURLTTL:      getEnvAsDuration("MEDIA_R2_READ_URL_TTL", 15*time.Minute),
			R2RequestTimeout:  getEnvAsDuration("MEDIA_R2_REQUEST_TIMEOUT", 4*time.Second),
			R2UploadTimeout:   getEnvAsDuration("MEDIA_R2_UPLOAD_TIMEOUT", 2*time.Minute),
		},
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvAsBool(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	boolVal, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}

	return boolVal
}

func getEnvAsDuration(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	if strings.HasSuffix(value, "d") {
		daysPart := strings.TrimSuffix(value, "d")
		days, err := strconv.Atoi(daysPart)
		if err != nil {
			return fallback
		}
		return time.Duration(days) * 24 * time.Hour
	}
	d, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return d
}

func getEnvAsSlice(key string, fallback []string) []string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parts := strings.Split(value, ",")
	var result []string
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	if len(result) == 0 {
		return fallback
	}
	return result
}

func getEnvAsInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	i, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return i
}

func getEnvAsInt64(key string, fallback int64) int64 {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}
