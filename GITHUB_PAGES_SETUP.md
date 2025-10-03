# GitHub Pages 部署指南

本文档说明如何将 VitePress 文档网站部署到 GitHub Pages。

## 🎯 快速部署

### 步骤 1: 推送代码到 GitHub

```bash
# 添加所有文件
git add .

# 提交更改
git commit -m "feat(docs): add VitePress documentation website"

# 推送到 GitHub
git push origin main
```

### 步骤 2: 启用 GitHub Pages

1. 访问你的 GitHub 仓库: https://github.com/rexleimo/agno-Go
2. 点击 **Settings**（设置）
3. 在左侧菜单点击 **Pages**
4. 在 **Build and deployment** 部分:
   - **Source**: 选择 **GitHub Actions** ✅
5. 保存（会自动保存）

### 步骤 3: 等待部署完成

1. 点击仓库顶部的 **Actions** 标签
2. 查看 "Deploy VitePress Docs to GitHub Pages" 工作流
3. 等待构建完成（通常 2-3 分钟）
4. 绿色 ✅ 表示成功

### 步骤 4: 访问你的网站

部署完成后，访问:

**https://rexleimo.github.io/agno-Go/**

🎉 恭喜！你的文档网站已经上线了！

## 📋 工作流说明

### 自动部署触发条件

GitHub Actions 会在以下情况自动部署:

- 推送到 `main` 分支
- 修改了 `website/` 目录中的文件
- 修改了 `package.json`
- 修改了 `.github/workflows/deploy-docs.yml`

### 手动触发部署

如果需要手动触发部署:

1. 访问 **Actions** 标签
2. 点击左侧 "Deploy VitePress Docs to GitHub Pages"
3. 点击右侧 **Run workflow** 按钮
4. 选择 `main` 分支
5. 点击绿色 **Run workflow** 按钮

## 🔧 故障排查

### 问题 1: Actions 标签中没有工作流

**原因**: 工作流文件还没有推送到 GitHub

**解决**:
```bash
git add .github/workflows/deploy-docs.yml
git commit -m "feat(ci): add GitHub Actions workflow for docs deployment"
git push origin main
```

### 问题 2: 构建失败 - "Dependencies lock file is not found"

**原因**: npm 缓存配置问题（已修复）

**解决**: 工作流已更新为使用 `npm install` 而不是 `npm ci`

### 问题 3: 部署成功但页面显示 404

**原因**: `base` 配置不正确

**检查**: 确保 `website/.vitepress/config.ts` 中:
```ts
base: '/agno-Go/',  // 必须与仓库名匹配！
```

### 问题 4: CSS 样式丢失或链接失效

**原因**: 同上，`base` 配置问题

**解决**:
1. 确认 `base: '/agno-Go/'`
2. 重新构建: `npm run docs:build`
3. 推送到 GitHub

### 问题 5: 构建成功但网站不更新

**等待**: GitHub Pages CDN 缓存需要 1-3 分钟更新

**强制刷新**:
- Chrome/Firefox: `Ctrl + Shift + R` (Mac: `Cmd + Shift + R`)
- Safari: `Cmd + Option + R`

### 问题 6: 权限错误 "Resource not accessible by integration"

**原因**: 工作流权限不足

**解决**: 已在工作流中配置:
```yaml
permissions:
  contents: read
  pages: write
  id-token: write
```

## 📝 工作流配置详解

### 完整的工作流文件

位置: `.github/workflows/deploy-docs.yml`

```yaml
name: Deploy VitePress Docs to GitHub Pages

on:
  push:
    branches:
      - main
    paths:
      - 'website/**'
      - 'package.json'
      - '.github/workflows/deploy-docs.yml'
  workflow_dispatch:  # 允许手动触发

permissions:
  contents: read
  pages: write
  id-token: write

concurrency:
  group: pages
  cancel-in-progress: false

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4
        with:
          fetch-depth: 0  # 获取完整历史（用于 lastUpdated）

      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '20'

      - name: Setup Pages
        uses: actions/configure-pages@v4

      - name: Install dependencies
        run: npm install

      - name: Build VitePress site
        run: npm run docs:build

      - name: Upload artifact
        uses: actions/upload-pages-artifact@v3
        with:
          path: website/.vitepress/dist

  deploy:
    environment:
      name: github-pages
      url: ${{ steps.deployment.outputs.page_url }}
    needs: build
    runs-on: ubuntu-latest
    steps:
      - name: Deploy to GitHub Pages
        id: deployment
        uses: actions/deploy-pages@v4
```

### 关键配置说明

1. **触发条件**:
   - `push` 到 `main` 分支
   - 仅当 `website/`、`package.json` 或工作流文件改变时
   - 支持 `workflow_dispatch` 手动触发

2. **Node.js 版本**:
   - 使用 Node.js 20（VitePress 要求 18+）

3. **构建产物**:
   - 上传 `website/.vitepress/dist/` 目录
   - 这是 VitePress 构建后的静态文件

4. **部署**:
   - 使用官方 `actions/deploy-pages@v4`
   - 自动处理所有部署细节

## 🔄 更新文档流程

### 日常更新文档

```bash
# 1. 编辑文档
nano website/guide/my-page.md

# 2. 本地预览（可选）
npm run docs:dev

# 3. 提交并推送
git add website/guide/my-page.md
git commit -m "docs: update my-page guide"
git push origin main

# 4. 等待自动部署（2-3分钟）
# 5. 访问 https://rexleimo.github.io/agno-Go/
```

### 批量更新

```bash
# 编辑多个文件后...
git add website/
git commit -m "docs: update multiple guide pages"
git push origin main
```

## 🌐 自定义域名（可选）

如果你有自定义域名（如 `docs.agno-go.com`）:

### 步骤 1: 添加 CNAME 记录

在你的 DNS 提供商添加:

```
Type: CNAME
Name: docs
Value: rexleimo.github.io
```

### 步骤 2: 配置 GitHub Pages

1. 在 **Settings → Pages**
2. 在 **Custom domain** 输入: `docs.agno-go.com`
3. 勾选 **Enforce HTTPS**

### 步骤 3: 更新 VitePress 配置

编辑 `website/.vitepress/config.ts`:

```ts
export default defineConfig({
  base: '/',  // 改为根路径
  // ...
})
```

### 步骤 4: 添加 CNAME 文件

创建 `website/public/CNAME`:

```
docs.agno-go.com
```

## 📊 监控部署状态

### 查看部署历史

1. 访问 **Actions** 标签
2. 点击任意工作流运行
3. 查看详细日志和错误信息

### 部署状态徽章

在 README.md 添加徽章:

```markdown
[![Docs Deploy](https://github.com/rexleimo/agno-Go/actions/workflows/deploy-docs.yml/badge.svg)](https://github.com/rexleimo/agno-Go/actions/workflows/deploy-docs.yml)
```

## ✅ 部署检查清单

在推送到生产环境前检查:

- [ ] 所有内部链接工作正常
- [ ] 代码示例已测试
- [ ] `base` 配置正确 (`/agno-Go/`)
- [ ] 图片资源在 `website/public/` 目录
- [ ] 本地构建成功 (`npm run docs:build`)
- [ ] 本地预览正常 (`npm run docs:preview`)

## 🆘 获取帮助

如果遇到问题:

1. **查看 Actions 日志**: 点击失败的工作流查看详细错误
2. **检查文档**: [VitePress 部署指南](https://vitepress.dev/guide/deploy)
3. **GitHub Issues**: [创建 Issue](https://github.com/rexleimo/agno-Go/issues)
4. **社区支持**: [GitHub Discussions](https://github.com/rexleimo/agno-Go/discussions)

## 📚 相关文档

- [DOCS_SETUP.md](DOCS_SETUP.md) - VitePress 本地开发指南
- [website/README.md](website/README.md) - VitePress 项目说明
- [VitePress 官方文档](https://vitepress.dev/)
- [GitHub Pages 文档](https://docs.github.com/en/pages)

---

**祝部署顺利！** 🚀

如果文档对你有帮助，请给项目一个 ⭐ Star！
