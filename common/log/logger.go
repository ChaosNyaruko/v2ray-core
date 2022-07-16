package log

import (
	"io"
	"log"
	"os"
	"regexp"
	"time"

	"github.com/mattn/go-isatty"
	"github.com/v2fly/v2ray-core/v5/common/platform"
	"github.com/v2fly/v2ray-core/v5/common/signal/done"
	"github.com/v2fly/v2ray-core/v5/common/signal/semaphore"
)

// Writer is the interface for writing logs.
type Writer interface {
	Write(string) error
	io.Closer
}

// WriterCreator is a function to create LogWriters.
type WriterCreator func() Writer

type generalLogger struct {
	creator WriterCreator
	buffer  chan Message
	access  *semaphore.Instance
	done    *done.Instance
}

// NewLogger returns a generic log handler that can handle all type of messages.
func NewLogger(logWriterCreator WriterCreator) Handler {
	return &generalLogger{
		creator: logWriterCreator,
		buffer:  make(chan Message, 16),
		access:  semaphore.New(1),
		done:    done.New(),
	}
}

func (l *generalLogger) run() {
	defer l.access.Signal()

	dataWritten := false
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	logger := l.creator()
	if logger == nil {
		return
	}
	defer logger.Close()

	for {
		select {
		case <-l.done.Wait():
			return
		case msg := <-l.buffer:
			logger.Write(msg.String() + platform.LineSeparator())
			dataWritten = true
		case <-ticker.C:
			if !dataWritten {
				return
			}
			dataWritten = false
		}
	}
}

func (l *generalLogger) Handle(msg Message) {
	select {
	case l.buffer <- msg:
	default:
	}

	select {
	case <-l.access.Wait():
		go l.run()
	default:
	}
}

func (l *generalLogger) Close() error {
	return l.done.Close()
}

type consoleLogWriter struct {
	logger           *log.Logger
	severityDetector *regexp.Regexp
	routingDetector  *regexp.Regexp
	isTerminal       bool
}

func (w *consoleLogWriter) Write(s string) error {
	if w.isTerminal {
		if routeIndex := w.routingDetector.FindStringSubmatchIndex(s); len(routeIndex) == 4 {
			b, e := routeIndex[2], routeIndex[3]
			r := NewBrush("Custom")(s[b:e])
			s = s[0:b] + r + s[e:]
		}
		matches := w.severityDetector.FindStringSubmatch(s)
		level := Severity_name[int32(Severity_Info)]
		if len(matches) >= 2 {
			level = matches[1]
		}
		s = NewBrush(level)(s)
	}
	w.logger.Print(s)
	return nil
}

func (w *consoleLogWriter) Close() error {
	return nil
}

type fileLogWriter struct {
	file   *os.File
	logger *log.Logger
}

func (w *fileLogWriter) Write(s string) error {
	w.logger.Print(s)
	return nil
}

func (w *fileLogWriter) Close() error {
	return w.file.Close()
}

func newConsoleLogWriter(w io.Writer) *consoleLogWriter {
	return &consoleLogWriter{
		isTerminal:       isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd()),
		logger:           log.New(os.Stdout, "", log.Ldate|log.Ltime),
		severityDetector: regexp.MustCompile(`^\[(Unknown|Info|Warning|Debug|Error)\]`),
		routingDetector:  regexp.MustCompile(`(?m):.*[^\[].*(\[.*?\])$`),
	}
}

// CreateStdoutLogWriter returns a LogWriterCreator that creates LogWriter for stdout.
func CreateStdoutLogWriter() WriterCreator {
	return func() Writer {
		return newConsoleLogWriter(os.Stdout)
	}
}

// CreateStderrLogWriter returns a LogWriterCreator that creates LogWriter for stderr.
func CreateStderrLogWriter() WriterCreator {
	return func() Writer {
		return newConsoleLogWriter(os.Stderr)
	}
}

// CreateFileLogWriter returns a LogWriterCreator that creates LogWriter for the given file.
func CreateFileLogWriter(path string) (WriterCreator, error) {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o600)
	if err != nil {
		return nil, err
	}
	file.Close()
	return func() Writer {
		file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o600)
		if err != nil {
			return nil
		}
		return &fileLogWriter{
			file:   file,
			logger: log.New(file, "", log.Ldate|log.Ltime),
		}
	}, nil
}

func init() {
	RegisterHandler(NewLogger(CreateStdoutLogWriter()))
}
