# Contributing to gAI

Thank you for your interest in contributing to gAI! This guide will help you get started and ensure a smooth collaboration.

## Table of Contents

- [Introduction](#introduction)
- [Reporting Issues](#reporting-issues)
- [Contributing Code](#contributing-code)
- [Coding Style](#coding-style)
- [Testing](#testing)
- [Pull Requests](#pull-requests)
- [Communication](#communication)
- [License](#license)

## Introduction

gAI is a command line tool for AI tasks, supporting multiple AI providers and various commands for chat, code analysis, project initialization, and more. Contributions are welcome to improve functionality, fix bugs, or enhance documentation.

## Reporting Issues

If you find a bug or have a feature request, please open an issue on the [GitHub repository](https://github.com/mkloubert/gai/issues). Provide as much detail as possible, including:

- Steps to reproduce the issue
- Expected behavior
- Actual behavior
- Environment details (OS, Go version, gAI version)

## Contributing Code

### Setup

1. Fork the repository on GitHub.
2. Clone your fork locally:

   ```bash
   git clone https://github.com/your-username/gai.git
   cd gai
   ```

3. Ensure you have Go 1.24.2 or higher installed.
4. Install dependencies:

   ```bash
   go mod download
   ```

### Development

- Follow the existing code structure and conventions.
- Use the `github.com/spf13/cobra` package for CLI commands.
- Use the MIT license header in all new files.

### Coding Style

- Follow Go idioms and formatting (`gofmt`).
- Write clear, maintainable, and well-documented code.
- Use descriptive names for variables and functions.

## Pull Requests

- Keep your pull requests focused on a single issue or feature.
- Provide a clear description of the changes and why they are needed.
- Rebase your branch on the latest `main` branch before submitting.
- Ensure your code passes all tests and linting checks.

## Communication

- Use GitHub issues and pull requests for discussions.
- Be respectful and constructive in all communications.

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

Thank you for helping improve gAI!
