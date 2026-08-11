---
title: Code Execution Sandbox
description: Execute untrusted agent code and shell commands in disposable, language-agnostic containers with no network, no privileges, bounded resources, and hard timeouts.
outline: deep
---

# Code Execution Sandbox

Giving an agent a `run_code` capability is one of the strongest grants in the
framework. HNO's code execution sandbox confines that grant to a disposable
container: a fresh filesystem, no network, no capabilities, a hard time
budget, and bounded memory, processes, and output. The sandbox is
**language-agnostic** — it never parses the payload. Language support is a
template concern: each runtime maps to an entrypoint inside a prebuilt image.

::: warning Scope of this feature
This is **execution isolation**, not a full operating-system sandbox. It
prevents the payload from reaching the host filesystem, the network, or host
processes, and it bounds resource usage. It does not protect against
kernel-level escapes: for workloads that need that boundary, run the whole
agent service inside a VM or use a microVM runtime such as Firecracker or
gVisor.
:::

## Why a bare `exec.Command` is not enough

A payload handed to a subprocess can do anything the host user can do:

- `rm -rf /` and other filesystem destruction;
- reading `/etc/shadow`, the environment, or mounted secrets;
- reverse shells and network exfiltration;
- fork bombs and memory exhaustion that take down the host.

Command allowlists and shell-metacharacter filters are defense in depth at
best and trivia to bypass at worst. The security boundary has to be the
process itself, not a check performed before launching it.

## Provider model

The executor is a small interface:

```go
type Executor interface {
    Run(ctx context.Context, spec Spec) (Result, error)
    Close() error
}
```

`NewExecutor` in default `auto` mode probes for a local container runtime and
**fails closed** when none is found — it never silently falls back to the host
or a billable Cloud service:

```
podman (preferred, rootless-capable) → docker → error
```

- **podman** is preferred: no daemon, rootless support, and disposable
  containers fit the run-once semantics of the sandbox.
- **docker** works when the host already runs it.
- A `local` executor exists for development only and must be constructed
  explicitly via `NewLocalExecutor`; production code should never use it.

### E2B Cloud

E2B is an explicit remote provider for teams that need disposable Linux VMs
without operating a container runtime. It uses E2B's public REST and Connect
APIs because E2B currently publishes JavaScript and Python SDKs, not a stable
Go SDK.

```go
executor, err := run.NewExecutor(run.Config{
    Backend: "e2b", // auto never starts a billable Cloud sandbox
    E2B: &run.E2BConfig{
        TemplateID:    "your-pinned-e2b-template",
        ResourceShell: "bash", // executable in this template; supports ulimit -v/-u
        // APIKey: os.Getenv("E2B_API_KEY"), // optional when env is set
    },
})
```

The provider requires `E2B_API_KEY` (or `E2BConfig.APIKey`) and a template
ID. It creates a fresh secure sandbox with internet disabled, uploads code as
a temporary file, executes it, then deletes the sandbox. `Workspace` mounts
are intentionally rejected in the first version; E2B volumes and file upload
need an explicit capability design before exposure.

### Agent tool

Wrap the executor in `CodeExecutionToolkit` and pass it wherever an HNO agent
accepts a `toolkit.Toolkit`. It exposes `run_code(runtime, code,
timeout_seconds?)`. A missing executor is a configuration error; it never
falls back to host execution. Failed programs return their bounded stdout,
stderr, exit code, and error text as a structured tool result so the agent can
correct the next attempt.

```go
executor, err := run.NewExecutor(config)
if err != nil { return err }
codeTools := run.NewToolkit(executor)
defer codeTools.Close()
// Add codeTools to the agent's toolkit list.
```

## Quick start

```go
import "github.com/rexleimo/agno-go/pkg/hno/tools/run"

executor, err := run.NewExecutor(run.Config{Image: "agent-runner:1"})
if err != nil {
    // fail closed: no provider, no execution
    return err
}
defer executor.Close()

result, err := executor.Run(ctx, run.Spec{
    Runtime: "python",
    Code:    "print('hello from the sandbox')",
    Timeout: 30 * time.Second,
})
```

The local container provider sends the payload on **stdin**, never in the
process argument list. E2B uploads it as a temporary sandbox file before
execution. Neither provider puts code in `ps` output or command metadata.

## Spec

| Field | Meaning | Default |
|---|---|---|
| `Runtime` | Template name: `python`, `node`, `shell` | required |
| `Code` | Payload sent to the entrypoint on stdin | required |
| `Timeout` | Hard deadline; the run is killed on expiry | 30s |
| `MemoryLimit` | Container memory ceiling (bytes) | 512 MiB |
| `PidLimit` | Process count ceiling (fork-bomb guard) | 128 |
| `Network` | `NetworkNone` only in this version | none |
| `Workspace` | Host directory mounted read-write at `/workspace` | unmounted |
| `Env` | Explicit `KEY=VALUE` allowlist; never secrets | empty |
| `OutputLimit` | Captured stdout/stderr cap per stream | 1 MiB |

## Fixed sandbox policy

Every invocation is a fresh container. The following flags are fixed policy
and cannot be relaxed through the API:

```
--network none                 no exfiltration, no reverse shells
--cap-drop ALL                 no privilege escalation
--security-opt no-new-privileges
--pids-limit <N>               fork-bomb guard
--memory <N> --memory-swap <N> memory ceiling, no swap escape
--read-only                    immutable root filesystem
--tmpfs /tmp:rw,size=64m       scratch space only
--cidfile <tmp>                reliable cleanup handle
```

A timed-out run is killed and its container force-removed by ID. Output is
capped per stream with a truncation marker, so output flooding cannot exhaust
memory. Nothing is ever `exec`'d into a shared container: state cannot leak
between runs.

## Runtime templates

| Runtime | Entrypoint | Notes |
|---|---|---|
| `python` | `python3 -` | dependencies baked into the image |
| `node` | `node -` | |
| `shell` | `bash -s` | last resort, still confined |

The base image (`agent-runner:1` by default, overridable via
`run.Config.Image`) is the only surface that grows with language demand:
compiled runtimes and system tools belong there, not in the sandbox core.

## Deployment prerequisites

- Local provider: `podman` (preferred) or `docker` on the host. For rootless
  podman, `unshare -Ur true` should succeed. Build a sandbox image with pinned
  runtime versions (never `latest`).
- E2B provider: an E2B Cloud API key and a pinned template ID. E2B runs the
  sandbox in its Linux VM infrastructure; no host root or Docker install is
  required. Its `ResourceShell` executable must exist in the template and
  support `ulimit -v` and `ulimit -u`.

## Security model

- **Fail closed:** unknown runtime, missing provider, or invalid spec →
  denied.
- **No network by default** and no host mounts unless `Workspace` is set.
- **No retained state:** `--rm`, fresh container per run.
- **Resource bounds:** memory, processes, time, output.
- **Least privilege:** all capabilities dropped, no new privileges, read-only
  rootfs.
- **Audit:** results record runtime, duration, exit code, and output size;
  full untrusted output is never logged beyond its cap.

For E2B, the provider additionally requests `secure: true` and
`allow_internet_access: false`. The template controls VM-level resource
limits; the process wrapper also applies the requested memory and PID ceilings
with `ulimit` before executing the runtime.

## Limits

- Kernel exploits: local container isolation shares the host kernel. For
  untrusted kernel-level workloads, select E2B Cloud's VM-backed provider or
  deploy behind your own VM/microVM runtime.
- Compile-and-run workflows (Go/Rust/C) need a file-based template and are not
  in the initial template set.
- No network-enabled templates yet; data must be pre-staged into the image or
  workspace.
