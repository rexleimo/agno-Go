---
title: 代码执行沙盒
description: 在一次性、语言无关的容器中执行不可信 agent 代码与 shell 命令,默认无网络、无特权、资源有界、硬超时。
outline: deep
---

# 代码执行沙盒

给 agent 一个 `run_code` 能力,是框架里最强的授权之一。HNO 的代码执行沙盒把这个授权限制在一次性的容器里:全新文件系统、无网络、无 capabilities、硬时间预算、内存/进程/输出有界。沙盒本身**语言无关**——它从不解析负载。语言支持是模板的事:每种 runtime 映射到预构建镜像里的一个入口。

::: warning 功能边界
这是**执行隔离**,不是完整的操作系统沙盒。它能防止负载触及宿主机文件系统、网络和宿主进程,并限制资源使用;它不防护内核级逃逸:需要这层边界的负载,请把整个 agent 服务放进 VM,或使用 Firecracker/gVisor 这类 microVM runtime。
:::

## 为什么裸 `exec.Command` 不够

交给子进程的负载可以做宿主机用户能做的一切:

- `rm -rf /` 等文件系统破坏;
- 读取 `/etc/shadow`、环境变量或挂载的密钥;
- 反弹 shell 与网络外传;
- fork bomb 与内存耗尽拖垮宿主机。

命令白名单和 shell 元字符过滤最多是纵深防御,绕过它们只是时间问题。安全边界必须是**进程本身**,而不是启动前的一次检查。

## Provider 模型

执行器是一个小接口:

```go
type Executor interface {
    Run(ctx context.Context, spec Spec) (Result, error)
    Close() error
}
```

`NewExecutor` 探测容器 runtime,**找不到就 fail closed**——绝不静默降级到宿主执行:

```
podman(首选,支持 rootless)→ docker → 报错
```

- **podman** 优先:无守护进程、支持 rootless、一次性容器正好契合沙盒"用完即抛"的语义。
- **docker** 在宿主已有时可用。
- `local` 执行器仅供开发,必须通过 `NewLocalExecutor` 显式构造;生产代码不应使用。

## 快速开始

```go
import "github.com/rexleimo/agno-go/pkg/hno/tools/run"

executor, err := run.NewExecutor(run.Config{Image: "agent-runner:1"})
if err != nil {
    // fail closed:没有 provider,不执行
    return err
}
defer executor.Close()

result, err := executor.Run(ctx, run.Spec{
    Runtime: "python",
    Code:    "print('hello from the sandbox')",
    Timeout: 30 * time.Second,
})
```

负载总是通过 **stdin** 传入,绝不进入进程参数列表,代码不会泄露到 `ps` 输出或容器元数据。

## Spec

| 字段 | 含义 | 默认 |
|---|---|---|
| `Runtime` | 模板名:`python`、`node`、`shell` | 必填 |
| `Code` | 通过 stdin 发给入口的负载 | 必填 |
| `Timeout` | 硬期限;超时即杀 | 30s |
| `MemoryLimit` | 容器内存上限(字节) | 512 MiB |
| `PidLimit` | 进程数上限(fork bomb 防护) | 128 |
| `Network` | 本版本仅 `NetworkNone` | none |
| `Workspace` | 只读/读写挂载到 `/workspace` 的宿主目录 | 不挂载 |
| `Env` | 显式 `KEY=VALUE` 白名单;绝不放密钥 | 空 |
| `OutputLimit` | 每路 stdout/stderr 捕获上限 | 1 MiB |

## 固定沙盒策略

每次调用都是全新容器。以下标志是固定策略,API 无法放宽:

```
--network none                 无外传、无反弹 shell
--cap-drop ALL                 无特权提升
--security-opt no-new-privileges
--pids-limit <N>               fork bomb 防护
--memory <N> --memory-swap <N> 内存上限,无 swap 逃逸
--read-only                    只读根文件系统
--tmpfs /tmp:rw,size=64m       仅临时空间
--cidfile <tmp>                可靠清理句柄
```

超时的运行会被杀死,并按 ID 强删容器。输出按流截断并带截断标记,输出洪泛无法耗尽内存。绝不 `exec` 进共享容器:状态无法在运行之间泄露。

## Runtime 模板

| Runtime | 入口 | 备注 |
|---|---|---|
| `python` | `python3 -` | 依赖打进镜像 |
| `node` | `node -` | |
| `shell` | `bash -s` | 最后手段,仍受限制 |

基础镜像(默认 `agent-runner:1`,可用 `run.Config.Image` 覆盖)是唯一随语言需求增长的表面:编译型 runtime 和系统工具属于那里,不属于沙盒核心。

## 部署前提

- 宿主机有 `podman`(首选)或 `docker`。
- rootless podman 需要非特权用户命名空间可用:`unshare -Ur true` 应成功。
- 构建包含所需 runtime 的沙盒镜像,固定到确定性标签(绝不用 `latest`)。

## 安全模型

- **Fail closed:** 未知 runtime、缺 provider、非法 spec → 拒绝。
- **默认无网络**,不设 `Workspace` 就不挂宿主目录。
- **无状态残留:** `--rm`,每次运行全新容器。
- **资源有界:** 内存、进程、时间、输出。
- **最小特权:** 丢弃全部 capabilities、禁止新特权、只读根文件系统。
- **审计:** 结果记录 runtime、时长、退出码、输出大小;完整不可信输出绝不超限记入日志。

## 局限

- 内核漏洞:容器隔离共享宿主内核。对不可信的内核级负载,请部署在 VM 或 microVM runtime 之后。
- 编译-运行工作流(Go/Rust/C)需要基于文件的模板,不在初版模板集内。
- 暂无联网模板;数据需预置进镜像或 workspace。
