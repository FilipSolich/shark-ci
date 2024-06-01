package worker

type Pipeline struct {
	Name  string   `yaml:"name"`
	Image string   `yaml:"image"`
	Cmds  []string `yaml:"cmds"`
}
