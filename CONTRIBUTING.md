# Contributing Guidelines (CONTRIBUTING.md)

Thank you for your interest in contributing to the `searxng-mcp-gateway` project, a part of the **Antigravity Agent Ecosystem** managed by **DoctorM & Ai / TheNovaNodes**.

## How to Contribute

### 1. Reporting Bugs
If you find a bug, please open an issue and include:
- A clear and descriptive title.
- Steps to reproduce the issue.
- Expected behavior vs. actual behavior.
- Your environment details (Python version, OS, etc.).

### 2. Suggesting Enhancements
We welcome feature requests! Please provide:
- A clear description of the proposed feature.
- Why it would be beneficial to the ecosystem.
- Any potential implementation ideas.

### 3. Submitting Pull Requests
1.  **Fork the repository** and create a feature branch (`git checkout -b feature/your-feature-name`).
2.  **Write code** adhering to our standard guidelines:
    -   Follow PEP 8 for Python code.
    -   Write clean, modular, and well-documented code.
    -   Ensure all comments and documentation are in high-quality English.
3.  **Test your code**:
    -   Write unit tests for new functionality.
    -   Ensure all existing tests pass (`python -m pytest tests/ -v`).
4.  **Update documentation**:
    -   If you modify configuration options, update `searxng_gateway/config.py`, `.env.example`, and `README.md`.
5.  **Commit your changes**:
    -   Use the Conventional Commits format (e.g., `feat(search): add new parameter`, `fix(config): resolve timeout bug`).
6.  **Create a Pull Request**:
    -   Provide a clear PR description outlining the Problem Statement, Root Cause (if fixing a bug), Solution Details, and Verification Results.

## Local Development
To set up your local development environment:
```bash
git clone https://github.com/TheNovaNodes/searxng-mcp-gateway.git
cd searxng-mcp-gateway
python3 -m venv venv
source venv/bin/activate
pip install -e .[dev]
```

Thank you for contributing to the Antigravity Agent Ecosystem!