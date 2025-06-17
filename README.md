# gAI

> A command line tool for AI based tasks written in [Go](https://go.dev).

## Usage

Move to project root and execute

```bash
go run . --help
```

in your terminal.

## Settings

### Environment variables

| Name                     | Description                                                                                                                                                    | Example / Default        |
| ------------------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------ |
| `GAI_BASE_URL`           | Custom base URL for API operations like AI chats.                                                                                                              | `http://localhost:11434` |
| `GAI_CONTEXT`            | Name of the current AI context.                                                                                                                                | `Project X`              |
| `GAI_DEFAULT_CHAT_MODEL` | The default provider and model for chat conversations.                                                                                                         | `openai:gpt-4.1-mini`    |
| `GAI_EDITOR`             | Command of custom editor to use.                                                                                                                               | `nano`                   |
| `GAI_INPUT_SEPARATOR`    | Custom separator string for input.                                                                                                                             | ` `                      |
| `GAI_INPUT_ORDER`        | Comma separated list of flags in what order to read string. Allowed values ares `args`/`a`, `editor`/`e` and `stdin`/`in`                                      | `a,in,e`                 |
| `GAI_MAX_TOKENS`         | Maximum number tokens to return / use.                                                                                                                         | `10000`                  |
| `GAI_TEMPERATURE`        | Temperature value for AI operations. Usually between `0` and `2`.                                                                                              | `0.3`                    |
| `GAI_TERMINAL_FORMATTER` | Default formatter for syntax highlighting in terminal. See [chroma project](https://github.com/alecthomas/chroma/tree/master/formatters) for more information. | `terminal16m`            |
| `GAI_TERMINAL_STYLE`     | Default style for syntax highlighting in terminal.. See [chroma project](https://github.com/alecthomas/chroma/tree/master/styles) for more information.        | `dracula`                |
| `OPENAI_API_KEY`         | The key for the [OpenAI API](https://help.openai.com/en/articles/4936850-where-do-i-find-my-openai-api-key).                                                   | `sk-proj-...xyz`         |
