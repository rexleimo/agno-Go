# SurrealDB Session Demo

This example shows how to plug **Agno-Go** session storage into a SurrealDB cluster.

## Prerequisites

- SurrealDB `v1.3+`
- Go `1.23` (matching the workspace `go.mod`)
- Environment variables for authentication (see below)

You can launch SurrealDB locally with Docker:

```bash
docker run --rm -p 8000:8000 \
  surrealdb/surrealdb:latest \
  start --log trace --user root --pass root file:/data/surreal.db
```

## Environment variables

| Variable | Default | Description |
|----------|---------|-------------|
| `SURREAL_URL` | `http://localhost:8000` | HTTP endpoint for SurrealDB |
| `SURREAL_NAMESPACE` | `demo` | Surreal namespace |
| `SURREAL_DATABASE` | `demo` | Surreal database |
| `SURREAL_USERNAME` | _required_ | Username (basic auth) |
| `SURREAL_PASSWORD` | _required_ | Password (basic auth) |

## Run the demo

```bash
cd cmd/examples/surreal_demo
go run .
```

Expected output:

```
🌱 creating session 4a2f… 
✅ session stored in SurrealDB
🔄 updating session state …
📚 listing sessions for agent demo-agent …
• 4a2f… (user=demo-user, updated=2024-10-18T00:00:00Z)
📊 fetching SurrealDB metrics …
total sessions: 1, active in 24h: 1, active in 1h: 1
🧹 deleting demo session …
✨ cleanup complete
```

The sample walks through:

1. Creating a session via `surreal.NewStorage`
2. Updating session metadata/state
3. Listing sessions for a given agent
4. Retrieving activity metrics
5. Deleting the demo session

Use this template to wire SurrealDB into your own AgentOS deployments. Feel free to swap the `StorageConfig` table name or extend the payload with your domain-specific fields.
