---
title: "Disposable Code Sandboxes: Running LLM-Generated Code Without Fear"
description: "HNO executes agent-generated code in disposable containers: no network, no capabilities, bounded resources, hard timeouts, and a language-agnostic template model."
date: 2026-08-11
lastUpdated: 2026-08-11
author: HNO Team
category: Security engineering
tags:
  - AI agents
  - agent security
  - sandbox
  - code execution
  - podman
  - Go
head:
  - - meta
    - name: keywords
      content: "AI agent code execution sandbox, disposable containers, podman rootless, LLM generated code, agent tool security, run code sandbox"
  - - meta
    - property: og:type
      content: article
  - - meta
    - property: og:title
      content: "Disposable Code Sandboxes: Running LLM-Generated Code Without Fear"
  - - meta
    - property: og:description
      content: "How HNO executes agent-generated code in disposable containers with no network, no capabilities, bounded resources, and hard timeouts."
  - - meta
    - property: article:published_time
      content: "2026-08-11T00:00:00Z"
  - - link
    - rel: canonical
    - href: https://hno.agno.dev/blog/code-execution-sandbox
---

# Disposable Code Sandboxes: Running LLM-Generated Code Without Fear

Agents write code. The interesting question is where that code runs.

Give a model a `run_code` tool backed by `exec.Command` and you have handed
the host to anyone who can phrase a prompt. A payload can delete the
filesystem, read `/etc/shadow`, open a reverse shell, or fork until the
machine stops responding. Allowlists of commands and blocked shell
metacharacters are not a boundary — they are a puzzle with well-known
solutions.

The boundary has to be the process itself. HNO's code execution sandbox
runs every payload in a **disposable container**: a fresh filesystem, no
network, no capabilities, a hard time budget, and bounded memory, processes,
and output. When the run finishes — or times out — the container is gone.

## Run-once semantics

The sandbox is deliberately hostile to state. Every invocation creates a
fresh container with `--rm`, so a payload cannot leave anything behind:
no dropped files, no listening sockets, no background processes. A timed-out
run is killed and force-removed by container ID, so even an unresponsive
payload cannot leak a zombie.

State cannot leak between runs by construction, because there is no shared
container to `exec` into. This is the same mental model as serverless
functions: immutable input, bounded output, nothing retained.

## A fixed policy, not a tunable one

Callers do not choose the security posture; they choose what code to run.
The container flags are fixed policy:

- `--network none` — no exfiltration, no reverse shells, no C2;
- `--cap-drop ALL` and `no-new-privileges` — no privilege escalation;
- `--pids-limit` — fork bombs die at the pid ceiling;
- `--memory` with `--memory-swap` equal — no swap escape;
- `--read-only` rootfs with a small tmpfs — nothing persists.

Local-container payloads reach the sandbox on **stdin**, never in the argument
list. E2B uploads a temporary sandbox file. Neither path leaks code into `ps`
output or command metadata.

## Language-agnostic by template

The sandbox core never parses code, which is what keeps it language-neutral.
A runtime template maps a language name to an entrypoint inside a prebuilt
image: `python3 -`, `node -`, or `bash -s`. The payload is piped to the
entrypoint's stdin. The template set — and the image with its baked-in
dependencies — is the only thing that grows as your agent's language needs
grow. Compiled runtimes, package ecosystems, and system tools belong in the
image, not in the sandbox core.

## Why podman

The executor probes `podman`, then `docker`, and fails closed if neither
exists — it never silently falls back to running payloads on the host.
podman is the preferred backend for three reasons:

1. **No daemon.** There is no long-lived privileged process to attack or
   restart; a container is just a process tree.
2. **Rootless by design.** With unprivileged user namespaces, a container
   escape lands in an unprivileged user, not in root. The network cost of
   rootless mode is irrelevant here because the sandbox is offline by
   default.
3. **Disposable by default.** `--rm` semantics match run-once execution.

For teams already invested in Docker, the same code path works: the CLI
surface is compatible.

## When the host cannot run containers: E2B Cloud

Some teams have no root access, cannot install Docker or podman, or want the
execution plane outside their application server. For them HNO also provides
an **explicit** E2B Cloud provider. It creates a fresh secure E2B Linux VM,
requests `allow_internet_access: false`, uploads the payload as a temporary
file, runs it with memory and PID ceilings, then deletes the VM.

The choice is deliberately explicit: `Backend: "e2b"` plus an API key and a
pinned template ID. The default `auto` mode will not spend Cloud quota because
than a stable Go SDK, so the adapter uses E2B's documented REST and Connect
protocol directly and is covered with an HTTP lifecycle test.

## What the tests prove

The integration suite (gated behind an environment variable) runs payloads
that should never succeed:

- `rm -rf /` — confined to the container filesystem;
- socket connects to the public internet — refused, `NETWORK-BLOCKED`;
- `sleep 300` with a two-second timeout — killed, flagged `TimedOut`;
- a hundred-thousand-character output flood — truncated with a marker.

None of them touch the host. That is the property that makes `run_code` a
capability worth granting.

## What the sandbox is not

The local container providers share the host kernel. A kernel exploit is out
of scope for that layer: workloads that demand a VM boundary can select the
E2B Cloud provider, or run the agent service behind Firecracker or gVisor.
The sandbox also does not solve prompt injection — a model can still be
tricked into *calling* the tool — it just guarantees that whatever the tool
executes cannot reach the host.

## Getting started

```go
executor, err := run.NewExecutor(run.Config{Image: "agent-runner:1"})
// fail closed: no provider, no execution

result, err := executor.Run(ctx, run.Spec{
    Runtime: "python",
    Code:    "print('hello from the sandbox')",
})
```

Build one pinned image with the runtimes your agents need, install podman,
and `run_code` becomes a tool you can hand to any model. The full design is
in [Code Execution Sandbox](/guide/code-execution-sandbox).
