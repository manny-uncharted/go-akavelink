# go-akavelink

🚀 A Go-based HTTP server that wraps the Akave SDK, exposing Akave APIs over REST. Previous version of this repo was a CLI wrapper around the Akave SDK refer to [akavelink](https://github.com/akave-ai/akavelink).

## Project Goals

- Provide a production-ready HTTP layer around Akave SDK
- Replace dependency on CLI-based wrappers
- Make it easy to integrate Akave storage into other systems via simple REST APIs

## Dev Setup

```bash
git clone https://github.com/akave-ai/go-akavelink
cd go-akavelink
go run ./cmd/server
```

Visit `http://localhost:8080/health` to verify it works.

## Project Structure

```
go-akavelink/
├── cmd/              # Main entrypoint
│   └── server/
│       └── main.go   # Starts HTTP server
├── internal/         # Internal logic, not exported
│   └── sdk/          # Wrapper around Akave SDK
├── pkg/              # Public packages (if needed)
├── docs/             # Architecture, design, etc.
├── go.mod            # Go module definition
├── README.md         # This file
├── CONTRIBUTING.md   # Guide for contributors
```

## Contributing

This repo is open to contributions! See [`CONTRIBUTING.md`](./CONTRIBUTING.md).

- Check the [issue tracker](https://github.com/akave-ai/go-akavelink/issues) for `good first issue` and `help wanted` labels.
- Follow the PR checklist and formatting conventions.
