package xlog

type LoggerConfig struct {
	Level            string `yaml:"level"` // trace | debug | info | warn | error | fatal | panic
	SaveLoggerAsFile bool   `yaml:"save_logger_as_file"`
	Directory        string `yaml:"directory"` // log file path = Director + ProjectName + LoggerName + .log
	ProjectName      string `yaml:"project_name"`
	LoggerName       string `yaml:"logger_name"`
	MaxSize          int    `yaml:"max_size"`
	MaxBackups       int    `yaml:"max_backups"`
}
