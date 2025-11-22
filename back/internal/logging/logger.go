package logging

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
	slogmulti "github.com/samber/slog-multi"
	slogjson "github.com/veqryn/slog-json"
)

// ginContextHandler wraps slog.Handler для автоматического извлечения полей из gin.Context
type ginContextHandler struct {
	handler slog.Handler
}

// newGinContextHandler создает новый handler с поддержкой gin.Context
func newGinContextHandler(h slog.Handler) *ginContextHandler {
	return &ginContextHandler{handler: h}
}

// Enabled проверяет, включен ли уровень логирования
func (h *ginContextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

// Handle обрабатывает запись лога
func (h *ginContextHandler) Handle(ctx context.Context, r slog.Record) error {
	// Ищем gin.Context в атрибутах
	var ginCtx *gin.Context
	newAttrs := make([]slog.Attr, 0, r.NumAttrs())

	r.Attrs(func(a slog.Attr) bool {
		// Проверяем, является ли значение gin.Context
		if c, ok := a.Value.Any().(*gin.Context); ok {
			ginCtx = c
			// Не добавляем сам контекст в атрибуты
			return true
		}
		newAttrs = append(newAttrs, a)
		return true
	})

	// Если нашли gin.Context, извлекаем поля
	if ginCtx != nil {
		if requestID, exists := ginCtx.Get("request_id"); exists {
			if rid, ok := requestID.(string); ok && rid != "" {
				newAttrs = append(newAttrs, slog.String("request_id", rid))
			}
		}
		if userID, exists := ginCtx.Get("user_id"); exists {
			if uid, ok := userID.(string); ok && uid != "" {
				newAttrs = append(newAttrs, slog.String("user_id", uid))
			}
		}
		newAttrs = append(newAttrs, slog.String("ip", ginCtx.ClientIP()))
		newAttrs = append(newAttrs, slog.String("method", ginCtx.Request.Method))
		newAttrs = append(newAttrs, slog.String("path", ginCtx.Request.URL.Path))
		newAttrs = append(newAttrs, slog.Int("status", ginCtx.Writer.Status()))

	}

	// Создаем новую запись с обновленными атрибутами
	newRecord := slog.NewRecord(r.Time, r.Level, r.Message, r.PC)
	newRecord.AddAttrs(newAttrs...)

	return h.handler.Handle(ctx, newRecord)
}

// WithAttrs добавляет атрибуты к handler
func (h *ginContextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &ginContextHandler{
		handler: h.handler.WithAttrs(attrs),
	}
}

// WithGroup добавляет группу к handler
func (h *ginContextHandler) WithGroup(name string) slog.Handler {
	return &ginContextHandler{
		handler: h.handler.WithGroup(name),
	}
}

// OutputFormat формат вывода логов
type OutputFormat string

const (
	FormatJSON OutputFormat = "json"
	FormatText OutputFormat = "text"
)

// OutputConfig конфигурация для одного вывода
type OutputConfig struct {
	Writer io.Writer
	Format OutputFormat
	Level  slog.Level
}

// LoggerConfig конфигурация логгера
type LoggerConfig struct {
	// Outputs список выводов (stdout, file, etc)
	Outputs []OutputConfig

	// DefaultAttrs дополнительные поля для всех логов (например service name)
	DefaultAttrs map[string]any

	// AddSource добавлять ли информацию о файле/строке
	AddSource bool
}

var globalLogger *slog.Logger

// InitGlobalLogger инициализирует глобальный логгер
func InitGlobalLogger(cfg LoggerConfig) {
	// Выводим конфигурацию логгера в консоль
	fmt.Println("=== Logger Configuration ===")
	fmt.Printf("AddSource: %v\n", cfg.AddSource)
	fmt.Printf("Outputs count: %d\n", len(cfg.Outputs))

	for i, output := range cfg.Outputs {
		writerType := "unknown"
		switch output.Writer {
		case os.Stdout:
			writerType = "stdout"
		case os.Stderr:
			writerType = "stderr"
		default:
			if _, ok := output.Writer.(*os.File); ok {
				writerType = "file"
			}
		}
		fmt.Printf("  Output #%d: writer=%s, format=%s, level=%s\n",
			i+1, writerType, output.Format, output.Level)
	}

	if len(cfg.DefaultAttrs) > 0 {
		fmt.Println("Default attributes:")
		for k, v := range cfg.DefaultAttrs {
			fmt.Printf("  %s: %v\n", k, v)
		}
	}
	fmt.Println("============================")

	handlers := make([]slog.Handler, 0, len(cfg.Outputs))

	for _, output := range cfg.Outputs {
		var handler slog.Handler

		switch output.Format {
		case FormatJSON:
			handler = slogjson.NewHandler(output.Writer, &slogjson.HandlerOptions{
				AddSource: cfg.AddSource,
				Level:     output.Level,
				JSONOptions: json.JoinOptions(
					json.Deterministic(true),
					jsontext.AllowDuplicateNames(true),
					jsontext.AllowInvalidUTF8(true),
					jsontext.EscapeForJS(true),
					jsontext.SpaceAfterColon(false),
					jsontext.SpaceAfterComma(true),
				),
			})
		case FormatText:
			handler = slog.NewTextHandler(output.Writer, &slog.HandlerOptions{
				AddSource: cfg.AddSource,
				Level:     output.Level,
			})
		default:
			handler = slog.NewTextHandler(output.Writer, &slog.HandlerOptions{
				AddSource: cfg.AddSource,
				Level:     output.Level,
			})
		}

		handlers = append(handlers, handler)
	}

	var finalHandler slog.Handler
	if len(handlers) == 1 {
		finalHandler = handlers[0]
	} else {
		finalHandler = slogmulti.Fanout(handlers...)
	}

	// Добавляем дефолтные атрибуты если они есть
	if len(cfg.DefaultAttrs) > 0 {
		attrs := make([]slog.Attr, 0, len(cfg.DefaultAttrs))
		for k, v := range cfg.DefaultAttrs {
			attrs = append(attrs, slog.Any(k, v))
		}
		finalHandler = finalHandler.WithAttrs(attrs)
	}

	// Оборачиваем в ginContextHandler для поддержки автоматического извлечения полей из gin.Context
	finalHandler = newGinContextHandler(finalHandler)

	globalLogger = slog.New(finalHandler)
	slog.SetDefault(globalLogger)
}

// OpenLogFile открывает файл для логов
func OpenLogFile(path string) (*os.File, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	return os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
}
