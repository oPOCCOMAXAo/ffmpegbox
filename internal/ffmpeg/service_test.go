package ffmpeg

import (
	"context"
	"strings"
	"testing"

	"github.com/opoccomaxao/ffmpegbox/internal/config"
	"github.com/opoccomaxao/ffmpegbox/internal/models"
)

func TestServiceValidateTask(t *testing.T) {
	cfg := &config.FFmpegConfig{
		BinaryPath:           "/usr/bin/ffmpeg",
		AllowedOutputFormats: []string{"mp4", "webm", "mkv", "avi", "mp3"},
		AllowedVideoCodecs:   []string{"libx264", "libx265", "libvpx", "copy"},
		AllowedAudioCodecs:   []string{"aac", "libmp3lame", "libopus", "copy"},
		AllowedPresets:       []string{"ultrafast", "fast", "medium", "slow"},
		MaxWidth:             3840,
		MaxHeight:            2160,
		MaxFramerate:         120,
	}

	svc := NewService(cfg)

	tests := []struct {
		name        string
		task        *models.Task
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid minimal task",
			task: &models.Task{
				OutputFormat: "mp4",
			},
			expectError: false,
		},
		{
			name: "valid complete task",
			task: &models.Task{
				OutputFormat: "mp4",
				VideoCodec:   "libx264",
				AudioCodec:   "aac",
				VideoBitrate: 2000000,
				AudioBitrate: 128000,
				Width:        1920,
				Height:       1080,
				Framerate:    30,
				Preset:       "medium",
			},
			expectError: false,
		},
		{
			name: "invalid - disallowed output format",
			task: &models.Task{
				OutputFormat: "flv",
			},
			expectError: true,
			errorMsg:    "output format",
		},
		{
			name: "invalid - disallowed video codec",
			task: &models.Task{
				OutputFormat: "mp4",
				VideoCodec:   "mpeg4",
			},
			expectError: true,
			errorMsg:    "video codec",
		},
		{
			name: "invalid - disallowed audio codec",
			task: &models.Task{
				OutputFormat: "mp4",
				AudioCodec:   "vorbis",
			},
			expectError: true,
			errorMsg:    "audio codec",
		},
		{
			name: "invalid - disallowed preset",
			task: &models.Task{
				OutputFormat: "mp4",
				Preset:       "veryslow",
			},
			expectError: true,
			errorMsg:    "preset",
		},
		{
			name: "invalid - negative width",
			task: &models.Task{
				OutputFormat: "mp4",
				Width:        -1920,
				Height:       1080,
			},
			expectError: true,
			errorMsg:    "resolution",
		},
		{
			name: "invalid - resolution exceeds max",
			task: &models.Task{
				OutputFormat: "mp4",
				Width:        7680,
				Height:       4320, // 8K
			},
			expectError: true,
			errorMsg:    "exceeds maximum",
		},
		{
			name: "invalid - framerate too high",
			task: &models.Task{
				OutputFormat: "mp4",
				Framerate:    240,
			},
			expectError: true,
			errorMsg:    "exceeds maximum",
		},
		{
			name: "invalid - negative video bitrate",
			task: &models.Task{
				OutputFormat: "mp4",
				VideoBitrate: -2000000,
			},
			expectError: true,
			errorMsg:    "bitrate",
		},
		{
			name: "invalid - negative audio bitrate",
			task: &models.Task{
				OutputFormat: "mp4",
				AudioBitrate: -128000,
			},
			expectError: true,
			errorMsg:    "bitrate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.ValidateTask(tt.task)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error, got nil")

					return
				}

				if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("error message %q does not contain %q", err.Error(), tt.errorMsg)
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestBuildCommandArgs(t *testing.T) {
	tests := []struct {
		name       string
		task       *models.Task
		inputPath  string
		outputPath string
		want       []string
	}{
		{
			name: "minimal task - format only",
			task: &models.Task{
				OutputFormat: "mp4",
			},
			inputPath:  "/tmp/input.avi",
			outputPath: "/tmp/output.mp4",
			want: []string{
				"-i", "/tmp/input.avi",
				"-f", "mp4",
				"-y",
				"/tmp/output.mp4",
			},
		},
		{
			name: "complete task with all parameters",
			task: &models.Task{
				OutputFormat: "mp4",
				VideoCodec:   "libx264",
				AudioCodec:   "aac",
				VideoBitrate: 2000000,
				AudioBitrate: 128000,
				Width:        1920,
				Height:       1080,
				Framerate:    30,
				Preset:       "medium",
			},
			inputPath:  "/tmp/input.avi",
			outputPath: "/tmp/output.mp4",
			want: []string{
				"-i", "/tmp/input.avi",
				"-c:v", "libx264",
				"-c:a", "aac",
				"-b:v", "2000000",
				"-b:a", "128000",
				"-s", "1920x1080",
				"-r", "30",
				"-preset", "medium",
				"-f", "mp4",
				"-y",
				"/tmp/output.mp4",
			},
		},
		{
			name: "task with video only",
			task: &models.Task{
				OutputFormat: "mp4",
				VideoCodec:   "libx264",
				VideoBitrate: 2000000,
				Width:        1920,
				Height:       1080,
			},
			inputPath:  "/tmp/input.avi",
			outputPath: "/tmp/output.mp4",
			want: []string{
				"-i", "/tmp/input.avi",
				"-c:v", "libx264",
				"-b:v", "2000000",
				"-s", "1920x1080",
				"-f", "mp4",
				"-y",
				"/tmp/output.mp4",
			},
		},
		{
			name: "task with audio only",
			task: &models.Task{
				OutputFormat: "mp3",
				AudioCodec:   "libmp3lame",
				AudioBitrate: 128000,
			},
			inputPath:  "/tmp/input.wav",
			outputPath: "/tmp/output.mp3",
			want: []string{
				"-i", "/tmp/input.wav",
				"-c:a", "libmp3lame",
				"-b:a", "128000",
				"-f", "mp3",
				"-y",
				"/tmp/output.mp3",
			},
		},
	}

	for _, tC := range tests {
		t.Run(tC.name, func(t *testing.T) {
			got := buildCommandArgs(tC.inputPath, tC.outputPath, tC.task)

			if len(got) != len(tC.want) {
				t.Errorf("buildCommandArgs() length = %d, want %d\nGot:  %v\nWant: %v", len(got), len(tC.want), got, tC.want)

				return
			}

			for i := range tC.want {
				if got[i] != tC.want[i] {
					t.Errorf("buildCommandArgs()[%d] = %q, want %q\nGot:  %v\nWant: %v", i, got[i], tC.want[i], got, tC.want)
				}
			}
		})
	}
}

func TestServiceBuildCommand(t *testing.T) {
	cfg := &config.FFmpegConfig{
		BinaryPath:           "/usr/bin/ffmpeg",
		AllowedOutputFormats: []string{"mp4", "webm"},
		AllowedVideoCodecs:   []string{"libx264"},
		AllowedAudioCodecs:   []string{"aac"},
		AllowedPresets:       []string{"medium"},
		MaxWidth:             3840,
		MaxHeight:            2160,
		MaxFramerate:         120,
	}

	svc := NewService(cfg)

	task := &models.Task{
		OutputFormat: "mp4",
		VideoCodec:   "libx264",
	}

	ctx := context.Background()
	cmd := svc.BuildCommand(ctx, "/tmp/input.avi", "/tmp/output.mp4", task)

	if cmd == nil {
		t.Fatal("BuildCommand() returned nil")
	}

	if len(cmd.Args) < 2 {
		t.Errorf("BuildCommand() returned command with too few args: %v", cmd.Args)
	}
}

func TestGenerateOutputFilename(t *testing.T) {
	cfg := &config.FFmpegConfig{
		BinaryPath:           "/usr/bin/ffmpeg",
		AllowedOutputFormats: []string{"mp4", "webm", "mp3"},
		AllowedVideoCodecs:   []string{"libx264"},
		AllowedAudioCodecs:   []string{"aac"},
		AllowedPresets:       []string{"medium"},
		MaxWidth:             3840,
		MaxHeight:            2160,
		MaxFramerate:         120,
	}

	svc := NewService(cfg)

	tests := []struct {
		name          string
		taskID        string
		inputFilename string
		task          *models.Task
		wantFilename  string
	}{
		{
			name:          "video to mp4",
			taskID:        "test-task-id",
			inputFilename: "video.avi",
			task: &models.Task{
				OutputFormat: "mp4",
			},
			wantFilename: "video-processed.mp4",
		},
		{
			name:          "audio to mp3",
			taskID:        "test-task-id",
			inputFilename: "audio.wav",
			task: &models.Task{
				OutputFormat: "mp3",
			},
			wantFilename: "audio-processed.mp3",
		},
		{
			name:          "no extension in input",
			taskID:        "test-task-id",
			inputFilename: "videofile",
			task: &models.Task{
				OutputFormat: "webm",
			},
			wantFilename: "videofile-processed.webm",
		},
	}

	for _, tC := range tests {
		t.Run(tC.name, func(t *testing.T) {
			got := svc.GenerateOutputFilename(tC.taskID, tC.inputFilename, tC.task)

			if got != tC.wantFilename {
				t.Errorf("GenerateOutputFilename() = %q, want %q", got, tC.wantFilename)
			}
		})
	}
}
