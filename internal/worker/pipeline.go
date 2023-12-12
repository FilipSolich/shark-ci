package worker

type Pipeline struct {
	Name  string         `yaml:"name"`
	Image string         `yaml:"image"`
	Jobs  map[string]Job `yaml:"jobs"`
}

type Job struct {
	Steps []Step `yaml:"steps"`
}

type Step struct {
	Name string   `yaml:"name"`
	Cmds []string `yaml:"cmds"`
}
