package ffmpeg

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/opoccomaxao/ffmpegbox/internal/config"
	"github.com/opoccomaxao/ffmpegbox/internal/models"
	"github.com/pkg/errors"
)

type Service struct {
	cfg *config.FFmpegConfig
}

func NewService(cfg *config.FFmpegConfig) *Service {
	return &Service{
		cfg: cfg,
	}
}

//nolint:cyclop,funlen // Validation logic is inherently complex but straightforward
func (s *Service) ValidateTask(task *models.Task) error {
	if !s.isAllowedOutputFormat(task.OutputFormat) {
		return errors.Wrapf(
			models.ErrInvalidParameter,
			"output format %q not allowed. Allowed: %v",
			task.OutputFormat,
			s.cfg.AllowedOutputFormats,
		)
	}

	if task.VideoCodec != "" {
		if !s.isAllowedVideoCodec(task.VideoCodec) {
			return errors.Wrapf(
				models.ErrInvalidParameter,
				"video codec %q not allowed. Allowed: %v",
				task.VideoCodec,
				s.cfg.AllowedVideoCodecs,
			)
		}
	}

	if task.AudioCodec != "" {
		if !s.isAllowedAudioCodec(task.AudioCodec) {
			return errors.Wrapf(
				models.ErrInvalidParameter,
				"audio codec %q not allowed. Allowed: %v",
				task.AudioCodec,
				s.cfg.AllowedAudioCodecs,
			)
		}
	}

	if task.Preset != "" {
		if !s.isAllowedPreset(task.Preset) {
			return errors.Wrapf(
				models.ErrInvalidParameter,
				"preset %q not allowed. Allowed: %v",
				task.Preset,
				s.cfg.AllowedPresets,
			)
		}
	}

	if task.Width > 0 || task.Height > 0 {
		if err := s.validateResolution(task.Width, task.Height); err != nil {
			return err
		}
	}

	if task.Framerate > 0 {
		if err := s.validateFramerate(task.Framerate); err != nil {
			return err
		}
	}

	if task.VideoBitrate != 0 {
		if err := s.validateBitrate(task.VideoBitrate); err != nil {
			return errors.Wrap(err, "video bitrate")
		}
	}

	if task.AudioBitrate != 0 {
		if err := s.validateBitrate(task.AudioBitrate); err != nil {
			return errors.Wrap(err, "audio bitrate")
		}
	}

	return nil
}

// BuildCommand constructs the ffmpeg command from validated task parameters.
// This method MUST only be called after ValidateTask has been called successfully.
func (s *Service) BuildCommand(ctx context.Context, inputPath, outputPath string, task *models.Task) *exec.Cmd {
	args := buildCommandArgs(inputPath, outputPath, task)

	// #nosec G204 -- All arguments are validated against whitelists before reaching this point
	return exec.CommandContext(ctx, s.cfg.BinaryPath, args...)
}

func buildCommandArgs(inputPath, outputPath string, task *models.Task) []string {
	args := []string{
		"-i", inputPath,
	}

	if task.VideoCodec != "" {
		args = append(args, "-c:v", task.VideoCodec)
	}

	if task.AudioCodec != "" {
		args = append(args, "-c:a", task.AudioCodec)
	}

	if task.VideoBitrate != 0 {
		args = append(args, "-b:v", strconv.FormatInt(task.VideoBitrate, 10))
	}

	if task.AudioBitrate != 0 {
		args = append(args, "-b:a", strconv.FormatInt(task.AudioBitrate, 10))
	}

	if task.Width > 0 && task.Height > 0 {
		args = append(args, "-s", fmt.Sprintf("%dx%d", task.Width, task.Height))
	}

	if task.Framerate > 0 {
		args = append(args, "-r", strconv.Itoa(task.Framerate))
	}

	if task.Preset != "" {
		args = append(args, "-preset", task.Preset)
	}

	args = append(args, "-f", task.OutputFormat)
	args = append(args, "-y")
	args = append(args, outputPath)

	return args
}

func (s *Service) GetVersion(ctx context.Context) (string, error) {
	// #nosec G204 -- Binary path is from config, -version is a safe static argument
	cmd := exec.CommandContext(ctx, s.cfg.BinaryPath, "-version")

	output, err := cmd.Output()
	if err != nil {
		return "", errors.Wrap(err, "failed to get ffmpeg version")
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) > 0 {
		return strings.TrimSpace(lines[0]), nil
	}

	return string(output), nil
}

func (s *Service) GenerateOutputFilename(taskID, inputFilename string, task *models.Task) string {
	baseName := strings.TrimSuffix(inputFilename, filepath.Ext(inputFilename))
	if baseName == "" {
		baseName = taskID
	}

	extension := task.OutputFormat

	return fmt.Sprintf("%s-processed.%s", baseName, extension)
}

func (s *Service) isAllowedOutputFormat(format string) bool {
	return slices.Contains(s.cfg.AllowedOutputFormats, format)
}

func (s *Service) isAllowedVideoCodec(codec string) bool {
	return slices.Contains(s.cfg.AllowedVideoCodecs, codec)
}

func (s *Service) isAllowedAudioCodec(codec string) bool {
	return slices.Contains(s.cfg.AllowedAudioCodecs, codec)
}

func (s *Service) isAllowedPreset(preset string) bool {
	return slices.Contains(s.cfg.AllowedPresets, preset)
}

func (s *Service) validateResolution(width, height int) error {
	if width <= 0 || height <= 0 {
		return errors.Wrapf(
			models.ErrInvalidParameter,
			"resolution dimensions must be positive, got %dx%d",
			width,
			height,
		)
	}

	if width > s.cfg.MaxWidth {
		return errors.Wrapf(
			models.ErrInvalidParameter,
			"width %d exceeds maximum %d",
			width,
			s.cfg.MaxWidth,
		)
	}

	if height > s.cfg.MaxHeight {
		return errors.Wrapf(
			models.ErrInvalidParameter,
			"height %d exceeds maximum %d",
			height,
			s.cfg.MaxHeight,
		)
	}

	return nil
}

func (s *Service) validateFramerate(framerate int) error {
	if framerate < 1 {
		return errors.Wrapf(
			models.ErrInvalidParameter,
			"framerate must be at least 1, got %d",
			framerate,
		)
	}

	if framerate > s.cfg.MaxFramerate {
		return errors.Wrapf(
			models.ErrInvalidParameter,
			"framerate %d exceeds maximum %d",
			framerate,
			s.cfg.MaxFramerate,
		)
	}

	return nil
}

func (s *Service) validateBitrate(bitrate int64) error {
	if bitrate <= 0 {
		return errors.Wrapf(
			models.ErrInvalidParameter,
			"bitrate must be positive, got %d",
			bitrate,
		)
	}

	return nil
}
