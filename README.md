# Jira CLI

This is a simple CLI tool for interacting with Jira.

## Features

*   View your assigned Jira tickets.
*   Open tickets in your browser.
*   Generate branch names from ticket information.

## Installation

1.  Clone the repository.
2.  Install the dependencies:

    ```bash
    go mod tidy
    ```

3.  Set the `JIRA_API_TOKEN` environment variable to your Jira API token.

## Usage

```bash
go run .
```

## Dependencies

*   [github.com/rivo/tview](https://github.com/rivo/tview)
*   [github.com/atotto/clipboard](https://github.com/atotto/clipboard)
*   [github.com/gdamore/tcell/v2](https://github.com/gdamore/tcell/v2)