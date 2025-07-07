# gAI - Command Line Tool for AI Tasks

> gAI is a versatile command line tool designed to interact with AI models for various tasks such as chatting, code analysis, code updating, project initialization, and more. It supports multiple AI providers and offers a rich set of commands and options to customize your AI interactions.

## Installation

### Prerequisites

- Go 1.24.2 or higher installed on your system.
- Internet connection for downloading dependencies and interacting with AI APIs.

### Building the Binary

1. Clone the repository:

   ```
   git clone https://github.com/mkloubert/gai.git
   cd gai
   ```

2. Build the binary using Go:

   ```
   go build -o gai .
   ```

3. Optionally, move the binary to a directory in your PATH:

   ```
   mv gai /usr/local/bin/
   ```

4. Verify the installation by running:
   ```
   gai --help
   ```

## Supported AI Providers

- **OpenAI**: Requires an API key set via `OPENAI_API_KEY` environment variable or `--api-key` flag.
- **Ollama**: Requires Ollama server running locally or accessible via configured base URL.

## Commands and Sub-Commands

### 1. `analize` (alias: `a`)

Analyze resources such as source code files.

#### Sub-commands:

- **`code` (alias: `c`)**

  Analyze source code files specified by `--file` or `--files` flags.

  **Usage:**

  ```
  gai analize code --file main.go --file utils.go "Explain the architecture of this code."
  ```

  **Description:**
  This command reads the content of the specified files, sends them to the AI for analysis, and returns detailed explanations. It supports multiple files and integrates their context for a comprehensive analysis.

- **`text` (aliases: `t`, `txt`)**

  Analyze text files specified by `--file` or `--files` flags.

  **Usage:**

  ```
  gai analize text --file document.txt "Summarize the key points."
  ```

  **Description:**
  This command reads text files, sends their content to the AI for analysis, and returns detailed explanations.

### 2. `chat` (alias: `c`)

Interact with AI via chat.

**Usage:**

```
gai chat "What is the weather today?"
```

**Options:**

- `--reset`, `-r`: Reset the conversation before starting.

**Description:**
Starts or continues a chat session with the AI. Supports sending files as context and resetting the conversation.

### 3. `commit`

Commit staged files with AI assistance.

**Usage:**

```
gai commit [options] [additional context]
```

**Description:**
This command analyzes the staged files in the git repository, optionally allows staging changed files, and generates a commit message following the Conventional Commits specification using AI. It supports retrying the commit message generation and confirms before committing.

### 4. `describe` (alias: `d`)

Describe resources such as images.

#### Sub-commands:

- **`images` (aliases: `image`, `img`, `imgs`, `i`)**

  Describe images with tags and detailed information.

  **Usage:**

  ```
  gai describe images --file photo.jpg "What is in this image?"
  ```

  **Description:**
  This command analyzes image files specified by `--file` or `--files` flags and generates a concise description, a short title, and a set of relevant tags for each image. It supports output in multiple languages and can store results in a database.

  **Flags:**

  - `--force-update`: Force update existing database entries.
  - `--max-tags`: Maximum number of tags to generate (default 10).
  - `--min-tags`: Minimum number of tags to generate (default 1).
  - `--update-existing`: Update existing database entries if present.

### 5. `init` (alias: `i`)

Initialize resources such as source code projects.

#### Sub-commands:

- **`code` (alias: `c`)**

  Initialize a new source code project with a given name and instructions.

  **Usage:**

  ```
  gai init code myproject "Create a new Go web server project."
  ```

  **Description:**
  This command creates a new project directory, generates multiple files and subfolders as needed, and provides a detailed README to get started quickly.

### 6. `list` (alias: `l`)

List various resources related to the app.

#### Sub-commands:

- **`conversation` (alias: `c`)**

  List the current conversation in the context.

  **Usage:**

  ```
  gai list conversation
  ```

  **Description:**
  Displays the current conversation history in the active context, showing roles and content with syntax highlighting.

- **`env` (alias: `e`)**

  List all environment variables used by the app.

  **Usage:**

  ```
  gai list env
  ```

  **Flags:**

  - `--no-sort`: Do not sort the environment variables.

  **Description:**
  Lists all environment variables loaded by the application, optionally sorted.

- **`files`**

  List files specified by `--file` or `--files`.

  **Usage:**

  ```
  gai list files --file main.go --files "*.go"
  ```

  **Flags:**

  - `--full`: Show full file paths.
  - `--with-types`: Show mime types of files.

  **Description:**
  Lists files matching the specified patterns or files, with options to show full paths and mime types.

- **`models`**

  List AI models for each supported provider.

  **Usage:**

  ```
  gai list models
  ```

  **Description:**
  Displays available AI models from configured providers such as OpenAI and Ollama.

### 7. `prompt` (alias: `p`)

Send a prompt to the AI.

**Usage:**

```
gai prompt "Write a poem about the sea."
```

**Description:**
Sends a single prompt to the AI and returns the response. Supports sending files as context.

### 8. `reset` (alias: `r`)

Reset resources.

#### Sub-commands:

- **`conversation` (alias: `c`)**

  Reset the current conversation context.

  **Usage:**

  ```
  gai reset conversation
  ```

  **Description:**
  Clears the current conversation history for the active context.

### 9. `update`

Update source code files as specified by `--file` or `--files` flags.

**Usage:**

```
gai update code --file main.go "Refactor this code to improve readability."
```

**Description:**
Sends the content of the specified files and a task description to the AI, which returns updated file contents along with explanations. The tool then writes the updates back to the files.

## Environment Variables

| Environment Variable     | CLI Flag(s)            | Description                                    | Example                            |
| ------------------------ | ---------------------- | ---------------------------------------------- | ---------------------------------- |
| `GAI_BASE_URL`           | `--base-url`, `-u`     | Custom base URL for AI API                     | `--base-url=https://api.custom`    |
| `GAI_CONTEXT`            | `--context`, `-c`      | Name of the current AI context                 | `--context=projectX`               |
| `GAI_DEFAULT_CHAT_MODEL` | `--model`, `-m`        | Default AI chat model (format: provider:model) | `--model=openai:gpt-4.1`           |
| `GAI_DATABASE`           | `--database`           | URI or path to database (usually SQLite)       | `--database=./images.db`           |
| `GAI_EDITOR`             | `--editor`             | Custom editor command                          | `--editor=vim`                     |
| `GAI_ENV_FILE`           | `--env-file`, `-e`     | Additional env files to load                   | `--env-file=.env.local`            |
| `GAI_FILE`               | `--file`, `-f`         | One or more files to use                       | `--file=main.go`                   |
| `GAI_FILES`              | `--files`              | One or more file patterns to use               | `--files=*.go`                     |
| `GAI_INPUT_ORDER`        |                        | Order of input sources: args, stdin, editor    | `args,stdin,editor`                |
| `GAI_INPUT_SEPARATOR`    |                        | Separator used when concatenating inputs       | `" "`                              |
| `GAI_MAX_TOKENS`         | `--max-tokens`         | Maximum number of tokens to use                | `--max-tokens=1000`                |
| `GAI_OUTPUT_FILE`        | `--output`, `-o`       | File to write output to                        | `--output=result.txt`              |
| `GAI_SCHEMA_FILE`        | `--schema`             | File with response format/schema               | `--schema=response.json`           |
| `GAI_SCHEMA_NAME`        | `--schema-name`        | Name of the response format/schema             | `--schema-name=MySchema`           |
| `GAI_SKIP_ENV_FILES`     | `--skip-env-files`     | Skip loading default `.env` files              | `--skip-env-files`                 |
| `GAI_SYSTEM_PROMPT`      | `--system`, `-s`       | Custom system prompt for AI                    | `--system="You are a helpful AI"`  |
| `GAI_SYSTEM_ROLE`        | `--system-role`        | Custom name/id of the system role              | `--system-role=system`             |
| `GAI_TEMP`               | `--temp`               | Custom temp folder                             | `--temp=./my-temp-folder`          |
| `GAI_TERMINAL_FORMATTER` | `--terminal-formatter` | Custom terminal formatter for output           | `--terminal-formatter=terminal16m` |
| `GAI_TERMINAL_STYLE`     | `--terminal-style`     | Custom terminal style for output               | `--terminal-style=dracula`         |
| `OPENAI_API_KEY`         | `--api-key`, `-k`      | API key for OpenAI provider                    | `OPENAI_API_KEY=sk-xxxx`           |

## Database Support and Usage

The tool supports SQLite databases for storing image descriptions and possibly other data.

- Configure the database path or URI using the `--database` flag or `GAI_DATABASE` environment variable.
- The database stores image metadata including file path, size, last modified time, title, description, and tags.

## Editor Integration

- Use the `--edit` flag to open your preferred text editor for input.
- Customize the editor command with the `--editor` flag or `GAI_EDITOR` environment variable.

## Conversation Storage and Management

- Conversations are stored locally in YAML files under the `.gai` directory in your home folder.
- Use `list conversation` to view the current conversation.
- Use `reset conversation` to clear the current conversation context.
- Context switching is supported via the `--context` flag or `GAI_CONTEXT` environment variable.

## Output Formatting and Highlighting

- Syntax highlighting is enabled by default when outputting to a terminal.
- Disable highlighting with the `--no-highlight` flag.
- Customize output appearance using `--terminal-formatter` and `--terminal-style` flags or corresponding environment variables.

## Input Sources and Order

- Input can be provided via command-line arguments, standard input, or an editor.
- Configure the order of input sources with the `GAI_INPUT_ORDER` environment variable (e.g., `args,stdin,editor`).
- Configure the separator used when concatenating inputs with the `GAI_INPUT_SEPARATOR` environment variable.

## Error Handling and Debugging

- Enable verbose/debug output with the `--verbose` flag.
- Debug logs provide detailed information about command execution and internal operations.

## Examples for All Commands

### Init Code

```bash
gai init code myproject "Create a new Go web server project."
```

### Describe Images with Database Usage

```bash
gai describe images --file photo.jpg --database ./images.db "What is in this image?"
```

### Update Code with File Updates

```bash
gai update code --file main.go "Refactor this code to improve readability."
```

### Prompt Command Usage with Files

```bash
gai prompt --file prompt.txt "Write a poem about the sea."
```

### List Subcommands Usage with Flags

```bash
gai list env --no-sort
```

```bash
gai list files --file main.go --with-types
```

## Supported File Types and Formats

- Images: JPEG, PNG, GIF, BMP, TIFF, WebP, HEIC, HEIF, AVIF
- Audio: MP3, WAV
- Documents: DOCX, PPTX, XLSX, PDF, HTML

## License and Contribution Guidelines

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

Contributions are welcome! Please open issues or pull requests on the [GitHub repository](https://github.com/mkloubert/gai).

## Known Issues and Limitations

- Some file formats may have limited support or require external dependencies.
- Ollama provider requires a running Ollama server.
- OpenAI usage requires a valid API key and may incur costs.

## Contact and Support

For support, issue reporting, or contact, please use the [GitHub repository](https://github.com/mkloubert/gai) issues section or contact Marcel Joachim Kloubert via GitHub.

---

For more information, visit the [GitHub repository](https://github.com/mkloubert/gai).
