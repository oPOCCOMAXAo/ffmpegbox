package models

type Task struct {
	ID             string     `gorm:"primaryKey;column:id;type:text"`
	CreatedAt      int64      `gorm:"column:created_at;not null;index;autoCreateTime"`
	ClientName     string     `gorm:"column:client_name;type:text;index:idx_tasks_client_status"`
	Status         TaskStatus `gorm:"column:status;not null;index;index:idx_tasks_client_status"`
	InputFilename  string     `gorm:"column:input_filename;not null;default:''"`
	OutputFilename string     `gorm:"column:output_filename;type:text"`
	ErrorMessage   string     `gorm:"column:error_message;type:text"`
	OutputFormat   string     `gorm:"column:output_format;not null;type:text"`
	VideoCodec     string     `gorm:"column:video_codec;type:text"`
	AudioCodec     string     `gorm:"column:audio_codec;type:text"`
	VideoBitrate   string     `gorm:"column:video_bitrate;type:text"`
	AudioBitrate   string     `gorm:"column:audio_bitrate;type:text"`
	Resolution     string     `gorm:"column:resolution;type:text"`
	Framerate      int        `gorm:"column:framerate;type:integer"`
	Preset         string     `gorm:"column:preset;type:text"`
}

func (Task) TableName() string {
	return "tasks"
}
