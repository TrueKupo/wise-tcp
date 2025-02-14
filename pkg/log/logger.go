package log

type Config struct {
	Prod bool   `yaml:"prod"`
	Name string `yaml:"name"`
}

type Logger interface {
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Debug(msg string, args ...interface{})
	Fatal(msg string, args ...interface{})
	Flush() error
}
