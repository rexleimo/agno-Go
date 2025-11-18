---
description: 根据现有的设计文档，将既有任务转换为可执行、按依赖排序的 GitHub issues。
tools: ['github/github-mcp-server/issue_write']
---

## 用户输入

```text
$ARGUMENTS
```

如有输入，你**必须**先参考用户输入后再继续。

## 执行概览

1. 在仓库根目录运行 `.specify/scripts/bash/check-prerequisites.sh --json --require-tasks --include-tasks`，解析输出中的 FEATURE_DIR 与 AVAILABLE_DOCS。所有路径必须为绝对路径。若参数包含单引号（如 "I'm Groot"），需用 `I'\''m Groot` 之类的转义写法；或在可行时改用双引号。
1. 从脚本输出中获取 **tasks** 文件的路径。
1. 运行以下命令获取 Git 远程地址：

```bash
git config --get remote.origin.url
```

**仅当远程地址指向 GitHub 时才能继续执行剩余步骤。**

1. 针对任务清单中的每个任务，调用 GitHub MCP 服务器在该远程仓库中创建新的 issue。

**在任何情况下，都不要向与当前远程 URL 不匹配的仓库创建 issue。**
