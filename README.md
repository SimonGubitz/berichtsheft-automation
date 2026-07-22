# Berichtsheft

![terminal preview](./assets/tui-screenshot.png)

Automatically generate weekly apprenticeship reports ("Berichtsheft") from your Jira activity using a local LLM (Ollama).

## Features

- Fetches your recent Jira activity
- Caches Jira API responses
- Generates a concise weekly summary using Ollama
- Supports both an interactive TUI and a simple CLI

## Prerequisites

- Go 1.26.5+ (if running locally)
- Docker & Docker Compose (optional)
- A Jira account
- An Ollama installation or the provided Docker setup

## Running

### Docker

```bash
docker compose up
```

### Local

```bash
make run tui
```

or

```bash
JIRA_URL=<username> \
JIRA_USERNAME=<username> \
JIRA_PASSWORD=<password> \
make run cli
```

## Purpose

Writing apprenticeship reports every week is repetitive.

This project collects your recent Jira activity, retrieves the relevant issues, and uses a local LLM to generate a short summary that can be used as the basis for your weekly Berichtsheft.

## Roadmap

- [x] Jira integration
- [x] Local caching
- [x] Ollama integration
- [ ] Terminal UI
- [ ] Docker Compose support
- [ ] PDF export
- [ ] Editing generated reports
- [ ] Support additional issue trackers
