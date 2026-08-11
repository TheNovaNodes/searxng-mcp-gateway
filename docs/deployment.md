# Deployment

This guide outlines the steps to configure and run the `searxng-mcp-gateway` either as a standalone Python process or integrated into a wider deployment.

## Installation and Start

The server requires Python 3.10 or higher. The recommended deployment method uses a Python virtual environment.

1. **Clone and Enter Directory**
   ```bash
   git clone https://github.com/TheNovaNodes/searxng-mcp-gateway.git
   cd searxng-mcp-gateway
   ```

2. **Set up Virtual Environment**
   ```bash
   python -m venv venv
   source venv/bin/activate
   ```

3. **Install Dependencies**
   To run the server:
   ```bash
   pip install -r requirements.txt
   ```
   For development and testing:
   ```bash
   pip install -e .[dev]
   ```

4. **Configuration**
   Copy the example environment configuration file to `.env`:
   ```bash
   cp .env.example .env
   ```
   Edit the `.env` file according to your target infrastructure (see Environment Variables table below).

5. **Start the Server**
   Since this is designed to be an MCP server, it can be executed via its main entry point to run using standard input/output transport:
   ```bash
   python -m searxng_gateway.server
   ```
   *Note*: When configured as an MCP client integration (like in Claude Desktop), the client handles executing this command.

## Environment Variables

The server behaves according to the variables defined in `.env` (or exported to the shell). Configuration parsing occurs in `searxng_gateway/config.py`.

| Variable | Description | Default |
|---|---|---|
| `SEARXNG_URL` | The URL of the target SearXNG instance endpoint. | `http://127.0.0.1:8081` |
| `DEFAULT_MAX_RESULTS` | Default number of search results to return if not specified. | 10 |
| `DEFAULT_LANGUAGE` | Default language for search results. | `auto` |
| `DEFAULT_SAFESEARCH` | Safesearch level (0=off, 1=moderate, 2=strict). | 0 |
| `DEFAULT_TIMEOUT` | Default request timeout to SearXNG in seconds. | 30 |
| `DEEP_RESEARCH_ORCHESTRATOR` | Path to the Deep Research orchestration shell/python script. | `/path/to/research-orchestrator.py` |
| `DEEP_RESEARCH_TIMEOUT` | Timeout limit for orchestrator script execution in seconds. | 60 |
| `SEMANTIC_ENABLED` | Toggle for semantic memory integration (`0`=disable, `1`=enable). | 0 |
| `SEMANTIC_TOP_K` | Number of results for semantic memory hybrid search. | 5 |
| `SEMANTIC_EXPAND` | Toggle context expansion in semantic memory. | 1 |
| `SEMANTIC_FUSION` | Method to fuse lexical/vector searches (`weighted`, `rrf`). | `weighted` |
| `MG_TRANSPORT` | Transport mechanism for the main MCP server process. | `stdio` |
| `MG_HOST` | Host binding for server running in HTTP/network mode. | `127.0.0.1` |
| `MG_PORT` | Port binding for server running in HTTP/network mode. | `8092` |

## Docker and Orchestration

### TODO
Currently, Docker deployment configurations (`Dockerfile`, `docker-compose.yml`) are not present in this repository.
When added, details regarding the Docker network bridging (ensuring communication between the MCP Gateway container and the SearXNG exit node container), volume definitions, and port mapping will be documented here.