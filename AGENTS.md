# AI Agent Guidelines (AGENTS.md)

Welcome to the `searxng-mcp-gateway` repository. If you are an AI Agent tasked with analyzing, modifying, or interacting with this codebase, please adhere to the following rules:

## 1. Project Context
*   **Ecosystem**: This project is part of the **Antigravity Agent Ecosystem** managed by **DoctorM & Ai / TheNovaNodes**.
*   **Purpose**: It is an MCP (Model Context Protocol) server for SearXNG metasearch engine integration.
*   **Philosophy**: The gateway is designed to return **raw data**. Do not add LLM-based synthesis to the search results returned by the MCP tools.

## 2. Code Quality & Standards
*   **Language**: All code comments, documentation, and commit messages MUST be in high-quality English.
*   **Style**: Adhere to standard PEP 8 conventions. Use type hints extensively.
*   **Testing**: All new features and bug fixes MUST be accompanied by unit tests. The test suite is powered by `pytest`. Run `python -m pytest tests/ -v` to verify your changes. Zero test failures are permitted.

## 3. Configuration Management
*   The central configuration is managed in `searxng_gateway/config.py` using environment variables.
*   **Synchronization**: Configuration keys and defaults MUST be kept synchronized across `searxng_gateway/config.py`, `README.md`, and `.env.example`.

## 4. Documentation
*   Technical documentation is maintained in the `/docs` directory (e.g., architecture.md, data-flow.md, deployment.md).
*   When documenting unimplemented features, use `### TODO` blocks to describe the missing parts instead of inventing functionality.

## 5. Development Setup
To set up the development environment and install dependencies, run:
```bash
pip install -e .[dev]
```