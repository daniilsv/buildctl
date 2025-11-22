package logging

import (
	"io"
	"log/slog"
	"os"
)

// LoggerBuilder билдер для создания конфигурации логгера
type LoggerBuilder struct {
	outputs      []OutputConfig
	defaultAttrs map[string]any
	addSource    bool
}

// NewLoggerBuilder создает новый билдер
func NewLoggerBuilder() *LoggerBuilder {
	return &LoggerBuilder{
		outputs:      make([]OutputConfig, 0),
		defaultAttrs: make(map[string]any),
		addSource:    false,
	}
}

// WithStdout добавляет вывод в stdout
func (b *LoggerBuilder) WithStdout(format OutputFormat, level slog.Level) *LoggerBuilder {
	b.outputs = append(b.outputs, OutputConfig{
		Writer: os.Stdout,
		Format: format,
		Level:  level,
	})
	return b
}

// WithFile добавляет вывод в файл
func (b *LoggerBuilder) WithFile(path string, format OutputFormat, level slog.Level) (*LoggerBuilder, error) {
	file, err := OpenLogFile(path)
	if err != nil {
		return b, err
	}

	b.outputs = append(b.outputs, OutputConfig{
		Writer: file,
		Format: format,
		Level:  level,
	})
	return b, nil
}

// WithWriter добавляет вывод с кастомным writer
func (b *LoggerBuilder) WithWriter(writer io.Writer, format OutputFormat, level slog.Level) *LoggerBuilder {
	b.outputs = append(b.outputs, OutputConfig{
		Writer: writer,
		Format: format,
		Level:  level,
	})
	return b
}

// WithServiceName добавляет service name в дефолтные атрибуты
func (b *LoggerBuilder) WithServiceName(name string) *LoggerBuilder {
	b.defaultAttrs["service"] = name
	return b
}

// WithAttr добавляет кастомный атрибут в дефолтные атрибуты
func (b *LoggerBuilder) WithAttr(key string, value any) *LoggerBuilder {
	b.defaultAttrs[key] = value
	return b
}

// WithAttrs добавляет несколько атрибутов в дефолтные атрибуты
func (b *LoggerBuilder) WithAttrs(attrs map[string]any) *LoggerBuilder {
	for k, v := range attrs {
		b.defaultAttrs[k] = v
	}
	return b
}

// WithSource включает добавление информации о файле/строке
func (b *LoggerBuilder) WithSource(addSource bool) *LoggerBuilder {
	b.addSource = addSource
	return b
}

// Build создает конфигурацию и инициализирует глобальный логгер
func (b *LoggerBuilder) Build() {
	cfg := LoggerConfig{
		Outputs:      b.outputs,
		DefaultAttrs: b.defaultAttrs,
		AddSource:    b.addSource,
	}
	InitGlobalLogger(cfg)
}

// BuildConfig создает только конфигурацию без инициализации
func (b *LoggerBuilder) BuildConfig() LoggerConfig {
	return LoggerConfig{
		Outputs:      b.outputs,
		DefaultAttrs: b.defaultAttrs,
		AddSource:    b.addSource,
	}
}
