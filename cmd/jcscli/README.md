# jcscli

`jcscli` is a command‑line tool for producing **canonical JSON** using the [JSON Canonicalization Scheme (JCS)](https://www.rfc-editor.org/rfc/rfc8785).  
It ensures that JSON data is serialized into a deterministic form suitable for hashing, signing, and comparison.

---

## Features

- Canonical JSON encoding (RFC 8785 compliant).
- Pretty‑print option for human‑friendly output.
- Quiet and verbose modes for controlling diagnostics.
- Interactive mode for typing/pasting JSON directly.
- Safe overwrite handling for output files.
- Clear exit codes for scripting.
- Short aliases for all flags.

---

## Installation

### From source (requires Go 1.21+):

```bash
go install github.com/Kbgjtn/jcs/cmd/jcscli@latest
```

This installs jcscli into your $GOPATH/bin or $HOME/go/bin.

From release binaries:

Download prebuilt binaries for Linux, macOS, and Windows from the Releases page.

## Usage Examples

Process a file and overwrite output:

```bash
jcscli -f input.json -o output.json -w
```

Pipe JSON from stdin and pretty‑print:

```bash
cat input.json | jcscli -p > output.json
```

Interactive mode (type/paste JSON, end with Ctrl+D):

```
jcscli -i
```

## Flags

```bash
Options:
-f, --file <path> Path to JSON input file (defaults to stdin)
-i, --interactive Read JSON interactively from stdin
-o, --output <path> Path to output file (defaults to stdout)
-w, --overwrite Allow overwriting existing output file
-p, --pretty Pretty-print the canonical JSON output
-q, --quiet Suppress non-fatal messages
-v, --verbose Print extra diagnostic information
-h, --help Show this help message
-V, --version Show program version
```
