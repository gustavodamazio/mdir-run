### Description:

**mdir-run** is a powerful Go-based command-line tool designed to execute a series of commands across multiple directories concurrently. It streamlines batch operations by automating tasks such as code updates, deployments, and maintenance across numerous projects. With real-time progress tracking and comprehensive logging, mdir-run enhances productivity and efficiency for developers and system administrators.

#### Key Features:

- **Concurrent Execution**: Run commands in multiple directories with configurable concurrency levels to optimize performance.
- **Interactive and Non-Interactive Modes**: Provide inputs via command-line flags or interactively through prompts, catering to different usage preferences.
- **Real-Time Progress Tracking**: Monitor the execution status of commands in each directory with live updates in the terminal.
- **Detailed Logging**: Generate log files capturing the execution status, output, and errors for auditing and debugging purposes.
- **Flexible Directory Selection**: Automatically detects and processes directories within a specified initial directory, with support for nested directories if needed.
- **Subdirectories entry points**: Insert subdirectories in relation to the parent, to execute commands in them if the directory exists, if it does not exist there will be no error.

#### Getting Started:

1. **Installation**:

   ```bash
   go install github.com/gustavodamazio/mdir-run@latest
   ```

2. **Usage**:

   You can run mdir-run using command-line flags or interactively.

   **Command-Line Flags**:

   ```bash
   mdir-run \
     -dir "/path/to/initial/directory" \
     -commands "git checkout dev; git pull; npm i --force; npm run deploy-dev" \
     -subdirs "functions" \
     -concurrency 5
   ```

   **Interactive Mode**:

   Simply run:

   ```bash
   mdir-run
   ```

   The tool will prompt you for the required inputs.

#### Examples:

- **Updating Multiple Git Repositories**:

  ```bash
  mdir-run -commands "git checkout dev; git pull"
  ```

#### Configuration Options:

- `-dir`: Specifies the initial directory containing subdirectories to process. If omitted, you will be prompted to enter it.
- `-commands`: A semicolon-separated list of commands to execute in each directory.
- `-concurrency`: The number of directories to process concurrently (default is 10).
- `-subdirs`: A semicolon-separated list of subdirectories to process in relation to the parent directory.

#### Logging:

Execution logs are saved in a `script.log` file within the initial directory, detailing the status, execution time, and any errors encountered.

#### Contributing:

Contributions are welcome! Please submit pull requests or open issues for any bugs or feature requests.