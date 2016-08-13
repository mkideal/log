package provider

type LoggerProvider interface {
	Write(level LogLevel, data []byte) error
}
