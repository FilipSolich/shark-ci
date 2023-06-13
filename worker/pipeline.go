package worker

type Pipeline struct {
	name      string
	baseImage string `yaml:"base_image"`
	jobs      map[string]Job
}

type Job struct {
	steps []Step
}

type Step struct {
	name string
	run  string
}
