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

| Name                     | Description                                                                                                                                                    | Example / Default     |
| ------------------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------- |
| `GAI_DEFAULT_CHAT_MODEL` | The default provider and model for chat conversations.                                                                                                         | `openai:gpt-4.1-mini` |
| `GAI_TERMINAL_FORMATTER` | Default formatter for syntax highlighting in terminal. See [chroma project](https://github.com/alecthomas/chroma/tree/master/formatters) for more information. | `terminal16m`         |
| `GAI_TERMINAL_STYLE`     | Default style for syntax highlighting in terminal.. See [chroma project](https://github.com/alecthomas/chroma/tree/master/styles) for more information.        | `dracula`             |
| `OPENAI_API_KEY`         | The key for the [OpenAI API](https://help.openai.com/en/articles/4936850-where-do-i-find-my-openai-api-key).                                                   | `sk-proj-...xyz`      |
