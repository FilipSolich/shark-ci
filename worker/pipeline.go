package worker

type Pipeline struct {
	Name      string
	BaseImage string `yaml:"base_image"`
	Jobs      map[string]Job
}

type Job struct {
	Steps []Step
}

type Step struct {
	Name string
	Run  string
}
