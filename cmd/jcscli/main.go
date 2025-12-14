package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/Kbgjtn/jcs"
	jsoniter "github.com/json-iterator/go"
)

var version = "0.0.1" // update as needed

func fatal(quiet bool, msg string, err error, code int) {
	if !quiet {
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", msg, err)
		} else {
			fmt.Fprintln(os.Stderr, msg)
		}
	}
	os.Exit(code)
}

func usage() {
	fmt.Fprintf(os.Stderr, `Usage: jcscli [options]

Options:
  -f, --file <path>       Path to JSON input file (defaults to stdin)
  -i, --interactive 		  Read JSON interactively from stdin
  -o, --output <path>     Path to output file (defaults to stdout)
  -w, --overwrite         Allow overwriting existing output file
  -p, --pretty            Pretty-print the canonical JSON output
  -q, --quiet             Suppress non-fatal messages
  -v, --verbose           Print extra diagnostic information
  -h, --help              Show this help message
  -V, --version           Show program version
`)
}

func main() {
	// Flags
	filePath := flag.String("file", "", "")
	flag.StringVar(filePath, "f", "", "")

	interactive := flag.Bool("interactive", false, "")
	flag.BoolVar(interactive, "i", false, "")

	outputPath := flag.String("output", "", "")
	flag.StringVar(outputPath, "o", "", "")

	overwrite := flag.Bool("overwrite", false, "")
	flag.BoolVar(overwrite, "w", false, "")

	pretty := flag.Bool("pretty", false, "")
	flag.BoolVar(pretty, "p", false, "")

	quiet := flag.Bool("quiet", false, "")
	flag.BoolVar(quiet, "q", false, "")

	verbose := flag.Bool("verbose", false, "")
	flag.BoolVar(verbose, "v", false, "")

	help := flag.Bool("help", false, "")
	flag.BoolVar(help, "h", false, "")

	showVersion := flag.Bool("version", false, "")
	flag.BoolVar(showVersion, "V", false, "")

	flag.Usage = usage
	flag.Parse()

	// Show help if requested
	if *help {
		usage()
		os.Exit(2)
	}

	if *showVersion {
		fmt.Printf("jcscli version %s\n", version)
		os.Exit(0)
	}

	// Detect "no args and stdin is a terminal"
	if !*interactive {
		fi, _ := os.Stdin.Stat()
		if *filePath == "" && flag.NFlag() == 0 && (fi.Mode()&os.ModeCharDevice) != 0 {
			usage()
			os.Exit(2)
		}
	}

	start := time.Now()

	// Choose input source
	var reader io.Reader
	var inputSize int64
	if *filePath != "" {
		f, err := os.Open(*filePath)
		if err != nil {
			fatal(*quiet, "Failed to open input file", err, 1)
		}
		defer f.Close()
		reader = f

		if fi, err := f.Stat(); err == nil {
			inputSize = fi.Size()
		}
	} else {
		reader = os.Stdin
	}

	// Decode JSON
	decodeStart := time.Now()
	var v any
	if err := jsoniter.ConfigCompatibleWithStandardLibrary.NewDecoder(reader).Decode(&v); err != nil {
		fatal(*quiet, "Invalid JSON", err, 1)
	}
	decodeElapsed := time.Since(decodeStart)

	// Pre‑allocate buffer
	bufCap := 1024 * 1024
	if inputSize > 0 {
		bufCap = int(inputSize) * 2
	}
	out := make([]byte, 0, bufCap)

	// Canonicalize with JCS
	encodeStart := time.Now()
	out, err := jcs.Append(out, v)
	if err != nil {
		fatal(*quiet, "Encoding error", err, 1)
	}
	encodeElapsed := time.Since(encodeStart)

	// Pretty-print if requested
	var final []byte
	if *pretty {
		var prettyBuf any
		if err := jsoniter.Unmarshal(out, &prettyBuf); err != nil {
			fatal(*quiet, "Pretty-print error", err, 1)
		}
		final, err = jsoniter.MarshalIndent(prettyBuf, "", "  ")
		if err != nil {
			fatal(*quiet, "Pretty-print error", err, 1)
		}
	} else {
		final = out
	}

	// Write output
	writeStart := time.Now()
	if *outputPath != "" {
		if !*overwrite {
			if _, err := os.Stat(*outputPath); err == nil {
				if !*quiet {
					fatal(*quiet, fmt.Sprintf("Output file %s already exists. Use --overwrite/-w to replace it.", *outputPath), nil, 1)
				}
			}
		}
		if err := os.WriteFile(*outputPath, final, 0o644); err != nil {
			fatal(*quiet, "Failed to write output file", err, 1)
		}
	} else {
		fmt.Println(string(final))
	}
	writeElapsed := time.Since(writeStart)

	// Verbose diagnostics
	if *verbose {
		totalElapsed := time.Since(start)
		fmt.Fprintf(os.Stderr, "Diagnostics:\n")

		if inputSize > 0 {
			fmt.Fprintf(os.Stderr, "  Input size: %d bytes\n", inputSize)
		}

		fmt.Fprintf(os.Stderr, "  Buffer capacity: %d bytes\n", bufCap)
		fmt.Fprintf(os.Stderr, "  Decode time: %v\n", decodeElapsed)
		fmt.Fprintf(os.Stderr, "  Encode time: %v\n", encodeElapsed)
		fmt.Fprintf(os.Stderr, "  Write time: %v\n", writeElapsed)
		fmt.Fprintf(os.Stderr, "  Total time: %v\n", totalElapsed)
	}

	os.Exit(0)
}
