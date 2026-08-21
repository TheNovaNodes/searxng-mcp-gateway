import json

path = "/root/projects/TheNovaNodes/antigravity-telegram-agent/mcp_config.json"
with open(path, "r") as f:
    config = json.load(f)

# Update searxng to use stdio instead of url
if "searxng" in config["servers"]:
    config["servers"]["searxng"] = {
      "name": "SearXNG Echelon Search Gateway",
      "type": "search",
      "plane": "data",
      "enabled": True,
      "command": "/root/projects/TheNovaNodes/searxng-mcp-gateway/.venv/bin/python",
      "args": [
        "-m",
        "searxng_gateway.server"
      ],
      "env": {
        "SEARXNG_URL": "http://127.0.0.1:8889"
      }
    }

with open(path, "w") as f:
    json.dump(config, f, indent=2)

print("mcp_config.json updated!")
