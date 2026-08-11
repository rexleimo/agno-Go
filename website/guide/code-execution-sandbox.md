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

`NewExecutor` probes for a container runtime and **fails closed** when none
is found — it never silently falls back to running on the host:

```
podman (preferred, rootless-capable) → docker → error
```

- **podman** is preferred: no daemon, rootless support, and disposable
  containers fit the run-once semantics of the sandbox.
- **docker** works when the host already runs it.
- A `local` executor exists for development only and must be constructed
  explicitly via `NewLocalExecutor`; production code should never use it.

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

The payload always arrives on **stdin**, never in the process argument list,
so code cannot leak into `ps` output or container metadata.

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

- `podman` (preferred) or `docker` on the host.
- For rootless podman, unprivileged user namespaces must work:
  `unshare -Ur true` should succeed.
- A sandbox image with the runtimes you expose, built and pinned to a
  deterministic tag (never `latest`).

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

## Limits

- Kernel exploits: container isolation shares the host kernel. For untrusted
  kernel-level workloads, deploy behind a VM or microVM runtime.
- Compile-and-run workflows (Go/Rust/C) need a file-based template and are not
  in the initial template set.
- No network-enabled templates yet; data must be pre-staged into the image or
  workspace.
