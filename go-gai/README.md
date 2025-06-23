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

### 3. `init` (alias: `i`)

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

### 4. `list` (alias: `l`)

List various resources related to the app.

#### Sub-commands:

- **`conversation` (alias: `c`)**

  List the current conversation in the context.

  **Usage:**

  ```
  gai list conversation
  ```

- **`env` (alias: `e`)**

  List all environment variables used by the app.

  **Usage:**

  ```
  gai list env
  ```

  **Flags:**

  - `--no-sort`: Do not sort the environment variables.

- **`files`**

  List files specified by `--file` or `--files`.

  **Usage:**

  ```
  gai list files --file main.go --files "*.go"
  ```

  **Flags:**

  - `--full`: Show full file paths.

### 5. `prompt` (alias: `p`)

Send a prompt to the AI.

**Usage:**

```
gai prompt "Write a poem about the sea."
```

**Description:**
Sends a single prompt to the AI and returns the response. Supports sending files as context.

### 6. `reset` (alias: `r`)

Reset resources.

#### Sub-commands:

- **`conversation` (alias: `c`)**

  Reset the current conversation context.

  **Usage:**

  ```
  gai reset conversation
  ```

### 7. `update` (alias: `u`)

Update resources such as source code files.

#### Sub-commands:

- **`code` (alias: `c`)**

  Update source code files as specified by `--file` or `--files` flags.

  **Usage:**

  ```
  gai update code --file main.go "Refactor this code to improve readability."
  ```

  **Description:**
  Sends the content of the specified files and a task description to the AI, which returns updated file contents along with explanations. The tool then writes the updates back to the files.

## Supported Environment Variables and CLI Flags

| Environment Variable     | CLI Flag(s)            | Description                                    | Example                            |
| ------------------------ | ---------------------- | ---------------------------------------------- | ---------------------------------- |
| `GAI_API_KEY`            | `--api-key`, `-k`      | Global API key for AI provider                 | `--api-key=sk-xxxx`                |
| `GAI_BASE_URL`           | `--base-url`, `-u`     | Custom base URL for AI API                     | `--base-url=https://api.custom`    |
| `GAI_CONTEXT`            | `--context`, `-c`      | Name of the current AI context                 | `--context=projectX`               |
| `GAI_DEFAULT_CHAT_MODEL` | `--model`, `-m`        | Default AI chat model (format: provider:model) | `--model=openai:gpt-4.1`           |
| `GAI_EDITOR`             | `--editor`             | Custom editor command                          | `--editor=vim`                     |
| `GAI_ENV_FILE`           | `--env-file`, `-e`     | Additional env files to load                   | `--env-file=.env.local`            |
| `GAI_FILE`               | `--file`, `-f`         | One or more files to use                       | `--file=main.go`                   |
| `GAI_FILES`              | `--files`              | One or more file patterns to use               | `--files=*.go`                     |
| `GAI_INPUT_ORDER`        |                        | Order of input sources: args, stdin, editor    | `args,stdin,editor`                |
| `GAI_INPUT_SEPARATOR`    |                        | Separator used when concatenating inputs       | `" "`                              |
| `GAI_MAX_TOKENS`         | `--max-tokens`         | Maximum number of tokens to use                | `--max-tokens=1000`                |
| `GAI_OUTPUT_FILE`        | `--output`, `-o`       | File to write output to                        | `--output=result.txt`              |
| `GAI_SKIP_ENV_FILES`     | `--skip-env-files`     | Skip loading default `.env` files              | `--skip-env-files`                 |
| `GAI_SYSTEM_PROMPT`      | `--system`, `-s`       | Custom system prompt for AI                    | `--system="You are a helpful AI"`  |
| `GAI_SYSTEM_ROLE`        | `--system-role`        | Custom name/id of the system role              | `--system-role=system`             |
| `GAI_TERMINAL_FORMATTER` | `--terminal-formatter` | Custom terminal formatter for output           | `--terminal-formatter=terminal16m` |
| `GAI_TERMINAL_STYLE`     | `--terminal-style`     | Custom terminal style for output               | `--terminal-style=dracula`         |
| `GAI_TEMPERATURE`        | `--temperature`, `-t`  | Temperature value for AI responses             | `--temperature=0.7`                |

## Supported Global CLI Flags

| Name                   | Description                          | Example                            |
| ---------------------- | ------------------------------------ | ---------------------------------- |
| `--api-key`, `-k`      | Global API key for AI provider       | `--api-key=sk-xxxx`                |
| `--base-url`, `-u`     | Custom base URL for AI API           | `--base-url=https://api.custom`    |
| `--context`, `-c`      | Name of the current AI context       | `--context=projectX`               |
| `--cwd`                | Current working directory            | `--cwd=/path/to/dir`               |
| `--edit`               | Open editor                          | `--edit`                           |
| `--editor`             | Custom editor command                | `--editor=vim`                     |
| `--eol`                | Custom EOL char sequence             | `--eol="\n"`                       |
| `--env-file`, `-e`     | One or more env files to load        | `--env-file=.env.local`            |
| `--file`, `-f`         | One or more files to use             | `--file=main.go`                   |
| `--files`              | One or more file patterns to use     | `--files=*.go`                     |
| `--home`               | User's home directory                | `--home=/home/user`                |
| `--skip-env-files`     | Do not load default .env files       | `--skip-env-files`                 |
| `--max-tokens`         | Maximum number of tokens             | `--max-tokens=1000`                |
| `--model`, `-m`        | Default chat model                   | `--model=openai:gpt-4.1`           |
| `--output`, `-o`       | Write output to this file            | `--output=result.txt`              |
| `--system`, `-s`       | Custom system prompt                 | `--system="You are a helpful AI"`  |
| `--system-role`        | Custom name/id of the system role    | `--system-role=system`             |
| `--temperature`, `-t`  | Custom temperature value             | `--temperature=0.7`                |
| `--terminal-formatter` | Custom terminal formatter for output | `--terminal-formatter=terminal16m` |
| `--terminal-style`     | Custom terminal style for output     | `--terminal-style=dracula`         |
| `--verbose`            | Verbose output                       | `--verbose`                        |

## Additional Notes

- The tool supports multiple AI providers, including OpenAI and Ollama.
- Conversations and context are stored locally in YAML files under the user's home directory.
- The tool supports syntax highlighting for output when run in a terminal.
- Editor integration allows editing prompts or inputs in your preferred text editor.
- Environment variables can be used to set default values for CLI flags.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

For more information, visit the [GitHub repository](https://github.com/mkloubert/gai).
