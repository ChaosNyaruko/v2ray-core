package log

import (
	"sync"

	"github.com/v2fly/v2ray-core/v5/common/serial"
)

// Message is the interface for all log messages.
type Message interface {
	String() string
}

// Handler is the interface for log handler.
type Handler interface {
	Handle(msg Message)
}

// Follower is the interface for following logs.
type Follower interface {
	AddFollower(func(msg Message))
	RemoveFollower(func(msg Message))
}

// GeneralMessage is a general log message that can contain all kind of content.
type GeneralMessage struct {
	Severity Severity
	Content  interface{}
}

// String implements Message.
func (m *GeneralMessage) String() string {
	return serial.Concat("[", m.Severity, "] ", m.Content)
}

// Record writes a message into log stream.
func Record(msg Message) {
	logHandler.Handle(msg)
}

var logHandler syncHandler

// RegisterHandler register a new handler as current log handler. Previous registered handler will be discarded.
func RegisterHandler(handler Handler) {
	if handler == nil {
		panic("Log handler is nil")
	}
	logHandler.Set(handler)
}

type syncHandler struct {
	sync.RWMutex
	Handler
}

func (h *syncHandler) Handle(msg Message) {
	h.RLock()
	defer h.RUnlock()

	if h.Handler != nil {
		h.Handler.Handle(msg)
	}
}

func (h *syncHandler) Set(handler Handler) {
	h.Lock()
	defer h.Unlock()

	h.Handler = handler
}

// TODO: using protobuf and auto generate
type Brush func(string) string

type colour = string

var reset colour = "\033[0m"

var colours = map[string]colour{
	"Unknown": "\033[0m",    // White
	"Error":   "\033[31m",   // Red
	"Warning": "\033[33m",   // Yellow
	"Info":    "\033[36m",   // Cyan
	"Debug":   "\033[34m",   // Blue
	"Custom":  "\033[1;33m", // Light Yellow
}

func NewBrush(i string) Brush {
	return func(s string) string {
		if len(s) == 0 {
			return ""
		}

		var end = len(s)
		if s[len(s)-1] == '\n' {
			if len(s) >= 2 && s[len(s)-2] == '\r' { // windows CRLF
				end = len(s) - 2
			} else {
				end = len(s) - 1
			}
		}
		return colours[i] + s[:end] + reset + s[end:]
	}
}
