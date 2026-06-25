package startup

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
)

type Endpoint struct {
	Name string
	URL  string
}

type StartupDisplay struct {
	AppName     string
	Version     string
	Env         string
	Config      map[string]string
	DB          string
	DBPoolStats string
	Server      string
	Endpoints   []Endpoint
	StartTime   time.Time
}

var (
	green      = color.New(color.FgGreen)
	yellow     = color.New(color.FgYellow)
	blue       = color.New(color.FgBlue)
	red        = color.New(color.FgRed)
	cyan       = color.New(color.FgCyan)
	magenta    = color.New(color.FgMagenta)
	dim        = color.New(color.Faint)
	bold       = color.New(color.Bold)
	boldGreen  = color.New(color.Bold, color.FgGreen)
	boldRed    = color.New(color.Bold, color.FgRed)
	boldYellow = color.New(color.Bold, color.FgYellow)
)

func ShowStartup(d *StartupDisplay) {
	start := time.Since(d.StartTime)

	fmt.Println()
	printDot(d.AppName+" v"+d.Version, green, "")
	printBadge(d.Env)
	fmt.Println()

	fmt.Println(dim.Sprint("─────────────────────────────"))

	printKeyValues(d.Config)
	printKeyValue("db", d.DB)
	if d.DBPoolStats != "" {
		printKeyValue("pool", d.DBPoolStats)
	}
	printKeyValue("server", d.Server)

	fmt.Println(dim.Sprint("─────────────────────────────"))

	for _, ep := range d.Endpoints {
		printDot(ep.Name, blue, "  "+ep.URL)
	}

	fmt.Println(dim.Sprint("─────────────────────────────"))

	printDot("ready in "+formatDuration(start), green, "")
	fmt.Println()
}

func printKeyValues(m map[string]string) {
	for k, v := range m {
		printKeyValue(k, v)
	}
}

func printKeyValue(key, value string) {
	dot := color.New(color.FgGreen).Sprint("●")
	label := padRight(key, 10)
	green.Printf("%s %s", dot, label)
	dim.Printf(": %s", value)
	fmt.Println()
}

func ShowConfigError(service string, err error) {
	fmt.Println()
	printDot(service+" error", red, "")
	fmt.Println()
	color.New(color.FgRed, color.Bold).Printf("  %v\n", err)
	fmt.Println()
	os.Exit(1)
}

func printDot(label string, c *color.Color, suffix string) {
	dot := color.New(color.FgGreen).Sprint("●")
	c.Print(dot + " " + label)
	if suffix != "" {
		dim.Print("  " + suffix)
	}
	fmt.Println()
}

func printBadge(env string) {
	switch env {
	case "development":
		yellow.Print("Vytora")
	case "production":
		green.Print("RUNNING")
	default:
		cyan.Print(env)
	}
}

func formatDuration(d time.Duration) string {
	if d < time.Microsecond {
		return d.Round(time.Nanosecond).String()
	}
	if d < time.Millisecond {
		return d.Round(time.Microsecond).String()
	}
	if d < time.Second {
		return d.Round(time.Millisecond).String()
	}
	return d.Round(time.Second).String()
}

func Shutdown() {
	fmt.Println()
	printDot("shutting down...", yellow, "")
	fmt.Println()
}

type ResponseRecorder struct {
	http.ResponseWriter
	status int
}

func NewResponseRecorder(w http.ResponseWriter) *ResponseRecorder {
	return &ResponseRecorder{ResponseWriter: w, status: http.StatusOK}
}

func (r *ResponseRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func LogRequest(w http.ResponseWriter, r *http.Request, start time.Time) {
	duration := time.Since(start)
	status := 200
	if rec, ok := w.(*ResponseRecorder); ok {
		status = rec.status
	}

	ms := duration.Round(time.Millisecond).Milliseconds()
	timeStr := time.Now().Format("15:04:05")
	methodStr := padRight(r.Method, 4)
	statusColored := padRight(colorizeStatus(status), 3)
	durStr := fmt.Sprintf("%3dms", ms)
	pathStr := padRight(r.URL.Path, 14)
	ipStr := padRight(r.RemoteAddr, 21)

	fmt.Printf("%s │ %s │ %s │ %s │ %s │ %s\n",
		timeStr,
		colorizeMethod(methodStr),
		statusColored,
		dim.Sprint(durStr),
		dim.Sprint(pathStr),
		dim.Sprint(ipStr),
	)
}

func padRight(s string, n int) string {
	if len(s) >= n {
		return s
	}
	return s + strings.Repeat(" ", n-len(s))
}

func colorizeMethod(method string) string {
	switch method {
	case "GET":
		return green.Sprint(method)
	case "POST":
		return yellow.Sprint(method)
	case "PUT":
		return blue.Sprint(method)
	case "DELETE":
		return red.Sprint(method)
	case "PATCH":
		return cyan.Sprint(method)
	default:
		return method
	}
}

func colorizeStatus(status int) string {
	s := strconv.Itoa(status)
	switch {
	case status >= 200 && status < 300:
		return boldGreen.Sprint(s)
	case status >= 300 && status < 400:
		return cyan.Sprint(s)
	case status >= 400 && status < 500:
		return boldYellow.Sprint(s)
	case status >= 500:
		return boldRed.Sprint(s)
	default:
		return s
	}
}
