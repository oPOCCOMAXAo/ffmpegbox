package config

import (
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/opoccomaxao/ffmpegbox/internal/models"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

const (
	bytesPerMB = 1024 * 1024
)

type Config struct {
	Server     ServerConfig     `yaml:"server"`
	Auth       AuthConfig       `yaml:"auth"`
	Processing ProcessingConfig `yaml:"processing"`
	FFmpeg     FFmpegConfig     `yaml:"ffmpeg"`
	Storage    StorageConfig    `yaml:"storage"`
	Logging    LoggingConfig    `yaml:"logging"`
}

type ServerConfig struct {
	Port         int    `yaml:"port"`
	ReadTimeout  string `yaml:"read_timeout"`
	WriteTimeout string `yaml:"write_timeout"`
	BindAddress  string `yaml:"bind_address"`
}

type AuthConfig struct {
	Enabled bool           `yaml:"enabled"`
	Clients []ClientConfig `yaml:"clients"`
}

type ClientConfig struct {
	APIKey           string `yaml:"api_key"`
	Name             string `yaml:"name"`
	MaxParallelTasks int    `yaml:"max_parallel_tasks"`
}

type ProcessingConfig struct {
	GlobalMaxParallelTasks int    `yaml:"global_max_parallel_tasks"`
	WorkerCount            int    `yaml:"worker_count"`
	MaxFileSizeMB          int    `yaml:"max_file_size_mb"`
	TaskTimeout            string `yaml:"task_timeout"`
	CleanupAge             string `yaml:"cleanup_age"`
}

type FFmpegConfig struct {
	BinaryPath           string   `yaml:"binary_path"`
	AllowedOutputFormats []string `yaml:"allowed_output_formats"`
	AllowedVideoCodecs   []string `yaml:"allowed_video_codecs"`
	AllowedAudioCodecs   []string `yaml:"allowed_audio_codecs"`
	AllowedPresets       []string `yaml:"allowed_presets"`
	MaxResolution        string   `yaml:"max_resolution"`
	MaxFramerate         int      `yaml:"max_framerate"`
}

type StorageConfig struct {
	TempDir      string `yaml:"temp_dir"`
	DatabasePath string `yaml:"database_path"`
}

type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

func Load(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read config file")
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, errors.Wrap(err, "failed to parse config YAML")
	}

	if err := cfg.Validate(); err != nil {
		return nil, errors.Wrap(err, "config validation failed")
	}

	return &cfg, nil
}

func (c *Config) Validate() error {
	if err := c.Server.Validate(); err != nil {
		return errors.Wrap(err, "server config")
	}

	if err := c.Auth.Validate(); err != nil {
		return errors.Wrap(err, "auth config")
	}

	if err := c.Processing.Validate(); err != nil {
		return errors.Wrap(err, "processing config")
	}

	if err := c.FFmpeg.Validate(); err != nil {
		return errors.Wrap(err, "ffmpeg config")
	}

	if err := c.Storage.Validate(); err != nil {
		return errors.Wrap(err, "storage config")
	}

	if err := c.Logging.Validate(); err != nil {
		return errors.Wrap(err, "logging config")
	}

	return nil
}

func (s *ServerConfig) Validate() error {
	if s.Port <= 0 || s.Port > 65535 {
		return errors.Wrapf(models.ErrInvalidParameter, "invalid port %d, must be 1-65535", s.Port)
	}

	if _, err := time.ParseDuration(s.ReadTimeout); err != nil {
		return errors.Wrap(err, "invalid read_timeout")
	}

	if _, err := time.ParseDuration(s.WriteTimeout); err != nil {
		return errors.Wrap(err, "invalid write_timeout")
	}

	if s.BindAddress == "" {
		return errors.Wrap(models.ErrInvalidParameter, "bind_address cannot be empty")
	}

	return nil
}

func (a *AuthConfig) Validate() error {
	if !a.Enabled {
		return nil
	}

	if len(a.Clients) == 0 {
		return errors.Wrap(models.ErrInvalidParameter, "auth is enabled but no clients configured")
	}

	seenKeys := make(map[string]bool)
	seenNames := make(map[string]bool)

	for i, client := range a.Clients {
		if client.APIKey == "" {
			return errors.Wrapf(models.ErrInvalidParameter, "client[%d]: api_key cannot be empty", i)
		}

		if seenKeys[client.APIKey] {
			return errors.Wrapf(models.ErrInvalidParameter, "client[%d]: duplicate api_key", i)
		}

		seenKeys[client.APIKey] = true

		if client.Name == "" {
			return errors.Wrapf(models.ErrInvalidParameter, "client[%d]: name cannot be empty", i)
		}

		if seenNames[client.Name] {
			return errors.Wrapf(models.ErrInvalidParameter, "client[%d]: duplicate name %q", i, client.Name)
		}

		seenNames[client.Name] = true

		if client.MaxParallelTasks < 1 {
			return errors.Wrapf(models.ErrInvalidParameter, "client[%d] (%s): max_parallel_tasks must be >= 1", i, client.Name)
		}
	}

	return nil
}

func (p *ProcessingConfig) Validate() error {
	if p.GlobalMaxParallelTasks < 1 {
		return errors.Wrap(models.ErrInvalidParameter, "global_max_parallel_tasks must be >= 1")
	}

	if p.WorkerCount < 1 {
		return errors.Wrap(models.ErrInvalidParameter, "worker_count must be >= 1")
	}

	if p.MaxFileSizeMB < 1 {
		return errors.Wrap(models.ErrInvalidParameter, "max_file_size_mb must be >= 1")
	}

	if _, err := time.ParseDuration(p.TaskTimeout); err != nil {
		return errors.Wrap(err, "invalid task_timeout")
	}

	if _, err := time.ParseDuration(p.CleanupAge); err != nil {
		return errors.Wrap(err, "invalid cleanup_age")
	}

	return nil
}

func (f *FFmpegConfig) Validate() error {
	if f.BinaryPath == "" {
		return errors.Wrap(models.ErrInvalidParameter, "binary_path cannot be empty")
	}

	if len(f.AllowedOutputFormats) == 0 {
		return errors.Wrap(models.ErrInvalidParameter, "allowed_output_formats cannot be empty")
	}

	if len(f.AllowedVideoCodecs) == 0 {
		return errors.Wrap(models.ErrInvalidParameter, "allowed_video_codecs cannot be empty")
	}

	if len(f.AllowedAudioCodecs) == 0 {
		return errors.Wrap(models.ErrInvalidParameter, "allowed_audio_codecs cannot be empty")
	}

	if len(f.AllowedPresets) == 0 {
		return errors.Wrap(models.ErrInvalidParameter, "allowed_presets cannot be empty")
	}

	if !isValidResolution(f.MaxResolution) {
		return errors.Wrapf(
			models.ErrInvalidParameter,
			"invalid max_resolution format: %q (expected WIDTHxHEIGHT)",
			f.MaxResolution,
		)
	}

	if f.MaxFramerate < 1 || f.MaxFramerate > 240 {
		return errors.Wrapf(models.ErrInvalidParameter, "max_framerate must be 1-240, got %d", f.MaxFramerate)
	}

	return nil
}

func (s *StorageConfig) Validate() error {
	if s.TempDir == "" {
		return errors.Wrap(models.ErrInvalidParameter, "temp_dir cannot be empty")
	}

	if s.DatabasePath == "" {
		return errors.Wrap(models.ErrInvalidParameter, "database_path cannot be empty")
	}

	return nil
}

func (l *LoggingConfig) Validate() error {
	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}

	if !validLevels[l.Level] {
		return errors.Wrapf(
			models.ErrInvalidParameter,
			"invalid log level %q, must be one of: debug, info, warn, error",
			l.Level,
		)
	}

	validFormats := map[string]bool{
		"json": true,
		"text": true,
	}

	if !validFormats[l.Format] {
		return errors.Wrapf(
			models.ErrInvalidParameter,
			"invalid log format %q, must be one of: json, text",
			l.Format,
		)
	}

	return nil
}

func (a *AuthConfig) GetClientByAPIKey(apiKey string) *ClientConfig {
	for i := range a.Clients {
		if a.Clients[i].APIKey == apiKey {
			return &a.Clients[i]
		}
	}

	return nil
}

func (s *ServerConfig) GetReadTimeout() time.Duration {
	d, _ := time.ParseDuration(s.ReadTimeout)

	return d
}

func (s *ServerConfig) GetWriteTimeout() time.Duration {
	d, _ := time.ParseDuration(s.WriteTimeout)

	return d
}

func (p *ProcessingConfig) GetMaxFileSizeBytes() int64 {
	return int64(p.MaxFileSizeMB) * bytesPerMB
}

func (p *ProcessingConfig) GetTaskTimeout() time.Duration {
	d, _ := time.ParseDuration(p.TaskTimeout)

	return d
}

func (p *ProcessingConfig) GetCleanupAge() time.Duration {
	d, _ := time.ParseDuration(p.CleanupAge)

	return d
}

func ParseResolution(resolution string) (int, int, error) {
	if !isValidResolution(resolution) {
		return 0, 0, errors.Wrapf(
			models.ErrInvalidParameter,
			"invalid resolution format: %q (expected WIDTHxHEIGHT)",
			resolution,
		)
	}

	parts := strings.Split(resolution, "x")

	width, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, errors.Wrap(err, "invalid width in resolution")
	}

	height, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, errors.Wrap(err, "invalid height in resolution")
	}

	if width < 1 || height < 1 {
		return 0, 0, errors.Wrap(models.ErrInvalidParameter, "resolution dimensions must be positive")
	}

	return width, height, nil
}

func isValidResolution(resolution string) bool {
	matched, _ := regexp.MatchString(`^\d+x\d+$`, resolution)

	return matched
}

func (f *FFmpegConfig) GetMaxResolutionPixels() (int, error) {
	width, height, err := ParseResolution(f.MaxResolution)
	if err != nil {
		return 0, err
	}

	return width * height, nil
}
