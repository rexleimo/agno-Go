---
title: "一次性代码沙盒:无惧运行 LLM 生成的代码"
description: "HNO 在一次性容器中执行 agent 生成的代码:无网络、无特权、资源有界、硬超时,配合语言无关的模板模型。"
date: 2026-08-11
lastUpdated: 2026-08-11
author: HNO Team
category: 安全工程
tags:
  - AI agents
  - agent 安全
  - 沙盒
  - 代码执行
  - podman
  - Go
head:
  - - meta
    - name: keywords
      content: "AI agent 代码执行沙盒, 一次性容器, podman rootless, LLM 生成代码, agent 工具安全, 代码运行沙盒"
  - - meta
    - property: og:type
      content: article
  - - meta
    - property: og:title
      content: "一次性代码沙盒:无惧运行 LLM 生成的代码"
  - - meta
    - property: og:description
      content: "HNO 如何在一次性容器中执行 agent 生成的代码:无网络、无特权、资源有界、硬超时。"
  - - meta
    - property: article:published_time
      content: "2026-08-11T00:00:00Z"
  - - link
    - rel: canonical
    - href: https://hno.agno.dev/zh/blog/code-execution-sandbox
---

# 一次性代码沙盒:无惧运行 LLM 生成的代码

Agent 会写代码。真正的问题在于:这些代码在哪里跑。

给模型一个基于 `exec.Command` 的 `run_code` 工具,等于把宿主机交给任何会写提示词的人。一段负载可以删掉文件系统、读取 `/etc/shadow`、打开反弹 shell,或者 fork 到机器宕机。命令白名单和 shell 元字符过滤不是边界——它们是一道有标准答案的谜题。

边界必须是进程本身。HNO 的代码执行沙盒把每个负载放进**一次性容器**:全新文件系统、无网络、无特权、硬时间预算、内存/进程/输出有界。运行结束——或超时——容器即消失。

## 用完即抛的语义

沙盒刻意与状态为敌。每次调用都用 `--rm` 创建全新容器,负载什么都留不下:没有落盘文件、没有监听端口、没有后台进程。超时的运行会被杀死并按容器 ID 强删,即使负载毫无响应也留不下僵尸进程。

由于不存在可 `exec` 的共享容器,状态在运行之间**构造性**地无法泄露。这和 serverless 函数的思维模型一致:输入不可变、输出有界、什么都不保留。

## 固定策略,而非可调参数

调用方选择的是"跑什么代码",不是"安全姿态"。容器标志是固定策略:

- `--network none` — 无外传、无反弹 shell、无 C2;
- `--cap-drop ALL` + `no-new-privileges` — 无特权提升;
- `--pids-limit` — fork bomb 死在进程数上限;
- `--memory` + 相等的 `--memory-swap` — 无 swap 逃逸;
- 只读根文件系统 + 小块 tmpfs — 什么都不持久化。

本地容器负载通过 **stdin** 进入沙盒,绝不进入参数列表。E2B 上传临时 sandbox 文件。两条路径都不会把代码泄露到 `ps` 输出或命令元数据。

## 语言无关:靠模板

沙盒核心从不解析代码,这正是它保持语言中立的原因。runtime 模板把语言名映射到预构建镜像里的入口:`python3 -`、`node -` 或 `bash -s`,负载经 stdin 管道送入。随着 agent 语言需求的增长,只需要扩展模板集和镜像(依赖打进镜像),沙盒核心不用动。编译型 runtime、包生态、系统工具都属于镜像,不属于沙盒核心。

## 为什么选 podman

执行器探测 `podman` → `docker`,都没有就 fail closed——绝不静默降级到宿主执行。podman 是首选后端,三个原因:

1. **无守护进程。** 没有常驻特权进程可供攻击或重启;容器只是一棵进程树。
2. **rootless 天生支持。** 借助非特权用户命名空间,容器逃逸只会落到非特权用户,而不是 root。rootless 的网络开销在这里无关紧要——沙盒默认断网。
3. **默认用完即抛。** `--rm` 语义正好契合单次执行。

已投入 Docker 的团队同样能用:CLI 表面兼容,同一套代码路径。

## 无法运行容器时:E2B Cloud

有些团队没有 root 权限、不能安装 Docker/Podman,或者希望执行平面完全离开应用服务器。HNO 为此提供**显式** E2B Cloud provider:创建一个全新的安全 E2B Linux VM,请求 `allow_internet_access: false`,上传临时代码文件,以内存/PID 上限运行,最后销毁 VM。

选择必须显式:`Backend: "e2b"` 加 API key 和固定 template ID。默认 `auto` 模式不会因为本地缺二进制就消耗 Cloud 额度。E2B 当前维护 JavaScript/Python SDK,没有稳定 Go SDK;因此适配器直接使用其公开 REST/Connect 协议,并有 HTTP 生命周期测试覆盖。

## 测试证明了什么

集成测试套件(由环境变量门控)运行那些**本就不该成功**的负载:

- `rm -rf /` — 被限制在容器文件系统内;
- 连接公网 socket — 被拒绝,输出 `NETWORK-BLOCKED`;
- 两秒超时下的 `sleep 300` — 被杀,标记 `TimedOut`;
- 十万字符输出洪泛 — 截断并打标记。

没有一个碰到宿主机。正是这个性质,让 `run_code` 成为一个值得授予的能力。

## 沙盒不是什么

本地容器 provider 共享宿主内核。内核漏洞超出了这一层的能力范围:需要 VM 边界的负载可选择 E2B Cloud provider,或把整个 agent 服务放进 Firecracker/gVisor 之后。沙盒也不解决提示注入——模型仍可能被骗去*调用*工具——它只保证工具执行的内容够不到宿主机。

## 快速开始

```go
executor, err := run.NewExecutor(run.Config{Image: "agent-runner:1"})
// fail closed:没有 provider,不执行

result, err := executor.Run(ctx, run.Spec{
    Runtime: "python",
    Code:    "print('hello from the sandbox')",
})
```

构建一个固定标签、预装 agent 所需 runtime 的镜像,装好 podman,`run_code` 就成了可以放心交给任何模型的能力。完整设计见[代码执行沙盒](/zh/guide/code-execution-sandbox)。
