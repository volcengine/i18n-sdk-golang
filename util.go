package i18n

import (
	"log"
	"os"
	"strings"
	"sync"
)

// Logger abstracts the procedure to record internal states with 4 levels.
type Logger interface {
	// Debug writes a log with debug level based on the format and variables.
	Debug(format string, v ...interface{})

	// Info writes a log with info level based on the format and variables.
	Info(format string, v ...interface{})

	// Warn writes a log with warn level based on the format and variables.
	Warn(format string, v ...interface{})

	// Error writes a log with error level based on the format and variables.
	Error(format string, v ...interface{})
}

// DefaultLogger creates a logger which is used by default with the standard log
// library and output the data to stdout.
func DefaultLogger() Logger {
	l := log.New(os.Stdout, "starling-goclient: ", log.Ldate|log.Lmicroseconds|log.Lmsgprefix)
	return &logService{l}
}

type logService struct {
	logger *log.Logger
}

// Debug implements the `Logger` interface.
func (l *logService) Debug(format string, v ...interface{}) {
	l.logger.Printf("[DEBUG]"+format, v...)
}

// Info implements the `Logger` interface.
func (l *logService) Info(format string, v ...interface{}) {
	l.logger.Printf("[INFO]"+format, v...)
}

// Warn implements the `Logger` interface.
func (l *logService) Warn(format string, v ...interface{}) {
	l.logger.Printf("[WARN]"+format, v...)
}

// Error implements the `Logger` interface.
func (l *logService) Error(format string, v ...interface{}) {
	l.logger.Printf("[ERROR]"+format, v...)
}

// Metricer provides the metrics facility to monitor the current service with
// time-series data to inspect the SDK internal state.
type Metricer interface {
	// EmitCounter adds a point with the given name, value, prefix and tags.
	EmitCounter(name string, value interface{}, tags map[string]string)
}

// DefaultMetricer creates a metricer which is used by default with standard log
// library and output the data to stderr.
func DefaultMetricer() Metricer {
	l := log.New(os.Stderr, "starling-goclient-metrics: ", log.Ldate|log.Lmicroseconds)
	return &metricsService{&logService{l}}
}

type metricsService struct {
	logger Logger
}

// EmitCounter implements the `Metricer` interface.
func (m *metricsService) EmitCounter(name string, value interface{}, tags map[string]string) {
	if tags == nil {
		tags = make(map[string]string)
	}
	tags["lang"] = "go"
	tags["version"] = SDKVersion
	tags["platform"] = Platform
	tags["ip"] = LocalIP

	tagsArr := make([]string, 0, len(tags))
	for k, v := range tags {
		tagsArr = append(tagsArr, k+":"+v)
	}
	m.logger.Info("%s=%v[%s]", name, value, strings.Join(tagsArr, ","))
}

// Group represents a class of work and forms a namespace in which
// units of work can be executed with duplicate suppression.
type Group struct {
	mu sync.Mutex       // protects m
	m  map[string]*call // lazily initialized
}

// call is an in-flight or completed Do call.
type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

// Do executes and returns the results of the given function, making
// sure that only one execution is in-flight for a given key at a
// time. If a duplicate comes in, the duplicate caller waits for the
// original to complete and receives the same results.
func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err
	}
	c := new(call)
	c.wg.Add(1)
	g.m[key] = c
	g.mu.Unlock()

	c.val, c.err = fn()
	c.wg.Done()

	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()

	return c.val, c.err
}

func composeClientKey(project, namespace string) string {
	return "[" + project + "]$#$[" + namespace + "]"
}
