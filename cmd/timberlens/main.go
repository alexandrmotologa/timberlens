package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alexandrmotologa/timberlens/internal/buffer"
	"github.com/alexandrmotologa/timberlens/internal/cluster"
	"github.com/alexandrmotologa/timberlens/internal/ingest"
	"github.com/alexandrmotologa/timberlens/internal/model"
	"github.com/alexandrmotologa/timberlens/internal/parser"
	"github.com/alexandrmotologa/timberlens/internal/server"
	"github.com/alexandrmotologa/timberlens/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
)

var (
	webPort     = flag.Int("web-port", 9900, "Port for embedded Web Studio HTTP server")
	noTUI       = flag.Bool("no-tui", false, "Disable terminal TUI (headless web-only mode)")
	noWeb       = flag.Bool("no-web", false, "Disable Web Studio HTTP server")
	bufferSize  = flag.Int("buffer-size", 500000, "Maximum ring buffer capacity (up to 1,000,000)")
	filePath    = flag.String("file", "", "Continuous log file to tail")
	syslogAddr  = flag.String("syslog", "", "UDP syslog address to listen on (e.g. :1514)")
	formatRegex = flag.String("format-regex", "", "Custom named-group regex for parsing proprietary plain-text logs")
	demoMode    = flag.Bool("demo", false, "Run built-in multi-service demo traffic generator")
)

func main() {
	flag.Parse()

	// Check for "demo" subcommand
	if len(os.Args) > 1 && os.Args[1] == "demo" {
		*demoMode = true
	}

	// Initialize Core Components
	ring := buffer.NewRingBuffer(*bufferSize)
	broker := server.NewSSEBroker()
	clusters := cluster.NewSignatureTracker(1000)
	histogram := buffer.NewTimeHistogram(1, 60)

	// Central dispatch callback for incoming parsed records
	consumer := func(rec *model.LogRecord) {
		ring.Push(rec)
		broker.Broadcast(rec)
		clusters.Record(rec)
		histogram.Record(rec)
	}

	// Custom regex parser wrapper if flag is provided
	var regexParser *parser.RegexParser
	if *formatRegex != "" {
		rp, err := parser.NewRegexParser(*formatRegex)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid --format-regex: %v\n", err)
			os.Exit(1)
		}
		regexParser = rp
	}

	wrappedConsumer := func(rec *model.LogRecord) {
		if regexParser != nil {
			if customRec, ok := regexParser.Parse([]byte(rec.Raw), rec.ID); ok {
				consumer(customRec)
				return
			}
		}
		consumer(rec)
	}

	// 1. Start Web Studio Server
	var srv *server.Server
	if !*noWeb {
		srv = server.NewServer(ring, broker, clusters, histogram)
		srv.RegisterStaticHandlers()

		addr := fmt.Sprintf(":%d", *webPort)
		go func() {
			if err := http.ListenAndServe(addr, srv.Handler()); err != nil && err != http.ErrServerClosed {
				fmt.Fprintf(os.Stderr, "Web Studio server error: %v\n", err)
			}
		}()
	}

	// 2. Start Ingestion Sources
	var activeSources int

	// Check if stdin is piped
	isPipe := false
	stat, err := os.Stdin.Stat()
	if err == nil && (stat.Mode()&os.ModeCharDevice) == 0 {
		isPipe = true
	}

	if isPipe {
		activeSources++
		stdinReader := ingest.NewStdinReader(os.Stdin, wrappedConsumer)
		go func() {
			_ = stdinReader.Start()
		}()
	}

	if *filePath != "" {
		activeSources++
		fileReader := ingest.NewFileReader(*filePath, wrappedConsumer)
		go func() {
			_ = fileReader.Start()
		}()
	}

	if *syslogAddr != "" {
		activeSources++
		syslogListener := ingest.NewSyslogListener(*syslogAddr, wrappedConsumer)
		go func() {
			_ = syslogListener.Start()
		}()
	}

	if *demoMode || activeSources == 0 {
		// If no sources provided, launch synthetic demo generator
		demoGen := ingest.NewDemoGenerator(40*time.Millisecond, consumer)
		go demoGen.Start()
	}

	// 3. User Interface Mode
	if *noTUI {
		fmt.Printf("TimberLens running in headless mode.\n")
		if !*noWeb {
			fmt.Printf("Web Studio accessible at: http://localhost:%d\n", *webPort)
		}
		fmt.Printf("Press Ctrl+C to stop.\n")

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		return
	}

	// Launch Charm Bubble Tea TUI with Mouse Motion support
	app := tui.NewAppModel(ring)
	p := tea.NewProgram(app, tea.WithAltScreen(), tea.WithMouseCellMotion())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running TimberLens TUI: %v\n", err)
		os.Exit(1)
	}
}
