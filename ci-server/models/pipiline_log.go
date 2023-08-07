package models

import "time"

type PipelineLog struct {
	ID         int64      `json:"id"`
	Cmd        string     `json:"cmd"`
	ReturnCode *int       `json:"return_code,omitempty"`
	StartedAt  *time.Time `json:"started_at,omitempty"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
	PipelineID int64      `json:"pipeline_id"`
}

type PipelineLogLine struct {
	Line          int64  `json:"line"`
	File          string `json:"file"`
	Content       string `json:"content"`
	PipelineLogID int64  `json:"pipeline_log_id"`
}
