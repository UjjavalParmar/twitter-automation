package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	GeminiKey   string
	Model       string
	MaxTokens   int
	Temperature float32
	TopP        float32

	XApiKey       string
	XApiSecret    string
	XAccessToken  string
	XAccessSecret string

	TZ              string
	PostsPerDay     int
	PostWindowStart string
	PostWindowEnd   string

	ReplyScanInterval time.Duration
	ReplyMinLikes     int
	ReplyMinRetweets  int
	ReplyMaxPerScan   int

	Lang    string
	DataDir string
}

func mustInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		i, err := strconv.Atoi(v)
		if err != nil {
			log.Fatalf("env %s invalid int: %v", key, err)
		}
		return i
	}
	return def
}

func mustFloat32(key string, def float32) float32 {
	if v := os.Getenv(key); v != "" {
		f, err := strconv.ParseFloat(v, 32)
		if err != nil {
			log.Fatalf("env %s invalid float: %v", key, err)
		}
		return float32(f)
	}
	return def
}

func Load() *Config {
	_ = godotenv.Load()

	cfg := &Config{
		GeminiKey:   os.Getenv("GEMINI_API_KEY"),
		Model:       envOr("MODEL", "gemini-2.0-flash"),
		MaxTokens:   mustInt("MAX_TOKENS", 256),
		Temperature: mustFloat32("TEMPERATURE", 0.9),
		TopP:        mustFloat32("TOP_P", 0.9),

		XApiKey:       os.Getenv("X_API_KEY"),
		XApiSecret:    os.Getenv("X_API_SECRET"),
		XAccessToken:  os.Getenv("X_ACCESS_TOKEN"),
		XAccessSecret: os.Getenv("X_ACCESS_SECRET"),

		TZ:              envOr("TZ", "Asia/Kolkata"),
		PostsPerDay:     mustInt("POSTS_PER_DAY", 5),
		PostWindowStart: envOr("POST_WINDOW_START", "09:00"),
		PostWindowEnd:   envOr("POST_WINDOW_END", "22:00"),

		ReplyScanInterval: time.Duration(mustInt("REPLY_SCAN_INTERVAL_MIN", 60)) * time.Minute,
		ReplyMinLikes:     mustInt("REPLY_MIN_LIKES", 50),
		ReplyMinRetweets:  mustInt("REPLY_MIN_RETWEETS", 10),
		ReplyMaxPerScan:   mustInt("REPLY_MAX_PER_SCAN", 3),

		Lang:    envOr("LANG", "en"),
		DataDir: envOr("DATA_DIR", "./data"),
	}

	if cfg.GeminiKey == "" {
		log.Fatal("GEMINI_API_KEY required")
	}
	for _, k := range []string{"X_API_KEY", "X_API_SECRET", "X_ACCESS_TOKEN", "X_ACCESS_SECRET"} {
		if os.Getenv(k) == "" {
			log.Fatalf("%s required", k)
		}
	}
	return cfg
}

func envOr(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
