package config

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	App           AppConfig           `mapstructure:"app"`
	Tracking      TrackingConfig      `mapstructure:"tracking"`
	Storage       StorageConfig       `mapstructure:"storage"`
	Telegram      TelegramConfig      `mapstructure:"telegram"`
	Discord       DiscordConfig       `mapstructure:"discord"`
	CrossPlatform CrossPlatformConfig `mapstructure:"cross_platform"`
}

type AppConfig struct {
	LogLevel string `mapstructure:"log_level"`
}

type TrackingConfig struct {
	PollInterval        time.Duration `mapstructure:"poll_interval"`
	SeventeenTrackToken string        `mapstructure:"seventeen_track_token"`
	RequestDelay        time.Duration `mapstructure:"request_delay"`
}

type StorageConfig struct {
	Path string `mapstructure:"path"`
}

type TelegramConfig struct {
	Enabled        bool    `mapstructure:"enabled"`
	BotToken       string  `mapstructure:"bot_token"`
	AllowedUserIDs []int64 `mapstructure:"allowed_user_ids"`
	AllowedChatIDs []int64 `mapstructure:"allowed_chat_ids"`
	PushChatIDs    []int64 `mapstructure:"push_chat_ids"`
}

type DiscordConfig struct {
	Enabled           bool     `mapstructure:"enabled"`
	BotToken          string   `mapstructure:"bot_token"`
	AllowedGuildIDs   []string `mapstructure:"allowed_guild_ids"`
	AllowedChannelIDs []string `mapstructure:"allowed_channel_ids"`
	PushChannelIDs    []string `mapstructure:"push_channel_ids"`
}

type MirrorRule struct {
	FromPlatform string `mapstructure:"from_platform"`
	FromChannel  string `mapstructure:"from_channel"`
	ToPlatform   string `mapstructure:"to_platform"`
	ToChannel    string `mapstructure:"to_channel"`
}

type CrossPlatformConfig struct {
	Enabled bool         `mapstructure:"enabled"`
	Mirrors []MirrorRule `mapstructure:"mirrors"`
}

func Load(cfgFile string) (*Config, error) {
	v := viper.New()

	setDefaults(v)
	configureEnvVars(v)

	if cfgFile != "" {
		v.SetConfigFile(cfgFile)
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("reading config file %s: %w", cfgFile, err)
		}
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("/etc/ems2sns")
		if err := v.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				return nil, fmt.Errorf("reading config: %w", err)
			}
			// No config file found — fine, rely on env vars and defaults
		}
	}

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("unmarshaling config: %w", err)
	}


	log.Printf("[config] log_level=%s", cfg.App.LogLevel)

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func setDefaults(v *viper.Viper) {
	// All keys MUST be registered here so AutomaticEnv can discover
	// the corresponding EMS2SNS_* environment variables.

	v.SetDefault("app.log_level", "info")

	v.SetDefault("tracking.poll_interval", "30m")
	v.SetDefault("tracking.request_delay", "2s")
	v.SetDefault("tracking.seventeen_track_token", "")

	v.SetDefault("storage.path", "data/subscriptions.json")

	v.SetDefault("telegram.enabled", false)
	v.SetDefault("telegram.bot_token", "")
	v.SetDefault("telegram.allowed_user_ids", []int64{})
	v.SetDefault("telegram.allowed_chat_ids", []int64{})
	v.SetDefault("telegram.push_chat_ids", []int64{})

	v.SetDefault("discord.enabled", false)
	v.SetDefault("discord.bot_token", "")
	v.SetDefault("discord.allowed_guild_ids", []string{})
	v.SetDefault("discord.allowed_channel_ids", []string{})
	v.SetDefault("discord.push_channel_ids", []string{})

	v.SetDefault("cross_platform.enabled", false)
	v.SetDefault("cross_platform.mirrors", []MirrorRule{})
}

func configureEnvVars(v *viper.Viper) {
	v.SetEnvPrefix("EMS2SNS")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
}

func (c *Config) validate() error {
	if !c.Telegram.Enabled && !c.Discord.Enabled {
		return fmt.Errorf("at least one notifier (telegram or discord) must be enabled")
	}
	if c.Telegram.Enabled && c.Telegram.BotToken == "" {
		return fmt.Errorf("telegram.bot_token is required when telegram is enabled")
	}
	if c.Discord.Enabled && c.Discord.BotToken == "" {
		return fmt.Errorf("discord.bot_token is required when discord is enabled")
	}
	if c.Tracking.PollInterval < 30*time.Second {
		return fmt.Errorf("tracking.poll_interval must be at least 30 seconds")
	}
	return nil
}
