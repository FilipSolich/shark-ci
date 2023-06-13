package ciserver

const (
	CIServer = "CI Server"

	EventHandlerPath = "/event_handler"

	JobsPath                    = "/jobs"
	JobsReportStatusHandlerPath = JobsPath + "/status"
	JobsPublishLogsHandlerPath  = JobsPath + "/logs"
)
