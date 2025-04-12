# mdir-run

A powerful Go-based command-line tool designed to execute a series of commands across multiple directories concurrently. It streamlines batch operations by automating tasks such as code updates, deployments, and maintenance across numerous projects.

## Key Features

- **Concurrent Execution**: Run commands in multiple directories with configurable concurrency levels to optimize performance.
- **Interactive and Non-Interactive Modes**: Provide inputs via command-line flags or interactively through prompts.
- **Real-Time Progress Tracking**: Monitor the execution status of commands in each directory with live updates in the terminal.
- **Comprehensive Logging System**: 
  - Generates a main `script.log` file with execution status for all operations
  - Creates individual detailed logs for successes and errors
  - Automatically archives logs into a compressed file (zip on Windows, tar.gz on Linux/macOS) at the end of execution
- **Flexible Directory Selection**: Automatically detects and processes directories within a specified initial directory.
- **Subdirectory Support**: Specify subdirectories to execute commands in (if they exist) without error if not found.
- **Robust Retry Mechanism**: Automatically retry failed commands a specified number of times with backoff.

## Installation

```bash
go install github.com/gustavodamazio/mdir-run@latest
```

## Usage

You can run mdir-run using command-line flags or interactively:

### Command-Line Mode

```bash
mdir-run \
  -dir "/path/to/initial/directory" \
  -commands "git checkout dev; git pull; npm i --force; npm run deploy-dev" \
  -subdirs "functions" \
  -concurrency 5 \
  -retries 3
```

### Interactive Mode

Simply run:

```bash
mdir-run
```

The tool will prompt you for the required inputs.

## Configuration Options

| Flag | Description | Default |
|------|-------------|---------|
| `-dir` | Specifies the initial directory containing subdirectories to process | (Required, prompted if omitted) |
| `-commands` | Semicolon-separated list of commands to execute in each directory | (Required, prompted if omitted) |
| `-concurrency` | Number of directories to process concurrently | 10 |
| `-subdirs` | Semicolon-separated list of subdirectories to process in relation to the parent directory | (None) |
| `-retries` | Number of retries for failed commands | 0 |

## Examples

### Updating Multiple Git Repositories

```bash
mdir-run -commands "git checkout dev; git pull"
```

### Building Multiple JavaScript Projects

```bash
mdir-run -dir "/path/to/projects" -commands "npm install; npm run build" -subdirs "frontend;backend" -concurrency 3
```

### Running Tests with Retry Capability

```bash
mdir-run -commands "go test ./..." -retries 2
```

## Logging System

The tool provides a comprehensive logging system:

1. **Main Log**: A `script.log` file is created in the initial directory, tracking status, execution time, and directory for each operation.

2. **Individual Logs**:
   - Success logs: `[directory_name]_success.txt` files contain detailed output from successful command executions.
   - Error logs: `[directory_name]_error.txt` files contain command output, error messages, and debugging information for failed executions.

3. **Log Archiving**: At the end of execution, all log files are automatically:
   - Archived into a single compressed file named `logs-[timestamp].zip` (Windows) or `logs-[timestamp].tar.gz` (Linux/macOS)
   - Original log files are deleted after successful archiving
   - The archive contains the main log and all individual success/error logs

This logging system provides both real-time monitoring and comprehensive post-execution analysis capabilities.

## Contributing

Contributions are welcome! Please submit pull requests or open issues for any bugs or feature requests.