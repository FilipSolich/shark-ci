package model2

type PipelineLog struct {
	ID         int64     `json:"id"`
	StartedAt  string    `json:"started_at"`
	FinishedAt string    `json:"finished_at"`
	Cmd        string    `json:"cmd"`
	Output     []LogLine `json:"output"`
	ReturnCode int       `json:"return_code"`
	PipelineID int64     `json:"pipeline_id"`
}

type LogLine struct {
	Line    int64  `json:"line"`
	File    string `json:"file"`
	Content string `json:"content"`
}
