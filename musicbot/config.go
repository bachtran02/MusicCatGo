package musicbot

import (
	"fmt"
	"os"

	"github.com/disgoorg/snowflake/v2"
	"gopkg.in/yaml.v3"
)

func ReadConfig(path string) (Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return Config{}, fmt.Errorf("failed to open config: %w", err)
	}
	defer file.Close()

	cfg := Config{
		MusicTracker: MusicTracker{
			Enabled: false,
		},
	}
	if err = yaml.NewDecoder(file).Decode(&cfg); err != nil {
		return Config{}, fmt.Errorf("failed to decode config: %w", err)
	}
	return cfg, nil
}

type Config struct {
	Bot          BotConfig    `yaml:"bot"`
	Nodes        []NodeConfig `yaml:"nodes"`
	MusicTracker MusicTracker `yaml:"music_tracker"`
}

type BotConfig struct {
	Token string `yaml:"token"`
}

type NodeConfig struct {
	Name      string `yaml:"name"`
	Address   string `yaml:"address"`
	Password  string `yaml:"password"`
	Secure    bool   `yaml:"secure"`
	SessionID string `yaml:"session_id"`
}

type MusicTracker struct {
	Enabled     bool         `yaml:"enabled"`
	ChannelID   snowflake.ID `yaml:"channel_id"`
	GuildID     snowflake.ID `yaml:"guild_id"`
	HttpPath    string       `yaml:"http_path"`
	HttpAddress string       `yaml:"http_address"`
}
