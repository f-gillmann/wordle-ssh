# wordle-ssh

A simple Wordle-like game that you can play over SSH.

## Building

### Prerequisites
- Go 1.25 or higher

### Build the project

```bash
make build
```

This will create the binary in the `bin/` directory.

### Run the project

```bash
make run
```

Or run the binary directly:

```bash
./bin/wordle-ssh
```

### Docker

Build the Docker image:

```bash
make docker-build
```

Run with Docker:

```bash
make docker-run
```
