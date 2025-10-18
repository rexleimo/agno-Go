# VitePress Documentation Setup and Deployment | VitePress 文档设置与部署

Complete guide for VitePress documentation website setup, local development, and GitHub Pages deployment.

完整的 VitePress 文档网站设置、本地开发和 GitHub Pages 部署指南。

---

## 🚀 Quick Deploy | 快速部署

### Three-Step Deployment | 三步部署

#### Step 1: Push to GitHub | 推送到 GitHub

```bash
git add .
git commit -m "feat(docs): add VitePress documentation website"
git push origin main
```

#### Step 2: Enable GitHub Pages | 启用 GitHub Pages

Visit | 访问: https://github.com/rexleimo/agno-Go/settings/pages

- **Source**: Select **GitHub Actions** ✅
- **源**: 选择 **GitHub Actions** ✅

#### Step 3: Access Website (wait 2-3 minutes) | 访问网站（等待 2-3 分钟）

🎉 **https://rexleimo.github.io/agno-Go/**

---

## 📁 Project Structure | 项目结构

```
agno-Go/
├── package.json              # Node.js dependencies | 依赖配置
├── website/                  # VitePress documentation | 文档目录
│   ├── .vitepress/
│   │   └── config.mjs       # Site configuration (ESM) | 站点配置（ESM）。如使用 TS，可用 config.ts
│   ├── index.md             # Homepage | 首页
│   ├── guide/               # User guides | 用户指南
│   │   ├── index.md         # What is Agno-Go? | 什么是 Agno-Go？
│   │   ├── quick-start.md   # 5-minute tutorial | 5分钟教程
│   │   ├── installation.md  # Setup guide | 安装指南
│   │   └── agent.md         # Agent guide | Agent 指南
│   ├── api/                 # API reference | API 参考
│   │   └── index.md         # API overview | API 总览
│   ├── advanced/            # Advanced topics | 高级主题
│   │   ├── architecture.md  # Architecture | 架构
│   │   ├── performance.md   # Benchmarks | 基准测试
│   │   └── deployment.md    # Deployment | 部署
│   └── examples/            # Examples | 示例
│       └── index.md         # Examples overview | 示例总览
└── .github/workflows/
    └── deploy-docs.yml      # Auto-deployment | 自动部署
```

---

## 🎉 Features | 功能特性

### Implemented Features | 已实现功能

- **🎨 Modern UI** - Beautiful Vue 3 based interface | 基于 Vue 3 的美观界面
- **🔍 Full-text Search** - Built-in local search | 内置本地搜索
- **🌙 Dark Mode** - Auto theme switching | 自动主题切换
- **📱 Responsive Design** - Mobile, tablet, desktop | 手机、平板、桌面适配
- **⚡ Fast Loading** - Vite powered | Vite 驱动
- **📖 Complete Navigation** - Top menu + sidebar | 顶部菜单 + 侧边栏
- **🔗 Source Links** - Jump to GitHub source | 跳转到 GitHub 源码
- **✏️ Edit Links** - "Edit on GitHub" button | "在 GitHub 上编辑"按钮
- **🕐 Last Updated** - Auto update time | 自动更新时间
- **🎯 Syntax Highlighting** - Go, Bash, YAML | Go、Bash、YAML 高亮

---

## 💻 Local Development | 本地开发

### Prerequisites | 前置要求

- Node.js 18+ ([Download](https://nodejs.org/))
- npm (comes with Node.js | 随 Node.js 安装)

### Quick Start | 快速开始

```bash
# Install dependencies | 安装依赖
npm install

# Start dev server | 启动开发服务器
npm run docs:dev
# Visit http://localhost:5173
```

**Permission Error Fix | 权限错误修复**:
```bash
# Fix npm cache ownership | 修复 npm 缓存权限
sudo chown -R $(whoami) ~/.npm

# Clean and reinstall | 清理并重新安装
npm cache clean --force
npm install
```

### Build Commands | 构建命令

```bash
# Build for production | 构建生产版本
npm run docs:build

# Preview production build | 预览生产构建
npm run docs:preview
```

---

## 📝 Editing Documentation | 编辑文档

### Add New Pages | 添加新页面

1. **Create markdown file | 创建 Markdown 文件**:
   ```bash
   touch website/guide/my-new-guide.md
   ```

2. **Edit config | 编辑配置** (`website/.vitepress/config.mjs` 或 `config.ts`):
   ```ts
   sidebar: {
     '/guide/': [
       {
         text: 'Guide',
         items: [
           { text: 'My New Guide', link: '/guide/my-new-guide' }
         ]
       }
     ]
   }
   ```

3. **Test locally | 本地测试**:
   ```bash
   npm run docs:dev
   ```

### Markdown Features | Markdown 功能

#### Code Blocks | 代码块

\`\`\`go
package main

func main() {
    fmt.Println("Hello!")
}
\`\`\`

#### Custom Containers | 自定义容器

::: tip
This is a tip | 这是提示
:::

::: warning
This is a warning | 这是警告
:::

::: danger STOP
Danger zone! | 危险区域！
:::

#### Line Highlighting | 行高亮

\`\`\`go{2,4-6}
package main

import "fmt" // highlighted | 高亮

func main() { // highlighted lines | 高亮行
    fmt.Println("Hello")
}
\`\`\`

---

## 🔧 Configuration | 配置

### Site Settings | 站点设置

Edit | 编辑 `website/.vitepress/config.mjs`（或 `config.ts`）:

```ts
export default defineConfig({
  title: "Agno-Go",                    // Site title | 站点标题
  description: "Your description",     // Meta description | 元描述
  base: '/agno-Go/',                   // Base URL (IMPORTANT!) | 基础 URL（重要！）

  themeConfig: {
    nav: [...],        // Top navigation | 顶部导航
    sidebar: {...},    // Sidebar menu | 侧边栏菜单
    socialLinks: [...] // GitHub, etc. | GitHub 等
  }
})
```

### Base URL Configuration | Base URL 配置

**IMPORTANT**: The `base` must match your repository name.

**重要**: `base` 必须与仓库名匹配。

```ts
base: '/agno-Go/',  // For https://rexleimo.github.io/agno-Go/
```

For custom domain | 自定义域名:

```ts
base: '/',  // For https://yourdomain.com/
```

---

## 🌐 GitHub Pages Deployment | GitHub Pages 部署

### Automatic Deployment | 自动部署

Documentation automatically deploys when you push to `main` branch.

推送到 `main` 分支时文档自动部署。

**Triggers | 触发条件**:
- Push to `main` branch | 推送到 `main` 分支
- Changes in `website/` directory | 修改 `website/` 目录
- Changes in `package.json` | 修改 `package.json`
- Manual trigger via Actions tab | 通过 Actions 标签手动触发

### GitHub Actions Workflow | GitHub Actions 工作流

Location | 位置: `.github/workflows/deploy-docs.yml`

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
  workflow_dispatch:  # Manual trigger | 手动触发

permissions:
  contents: read
  pages: write
  id-token: write

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4
        with:
          fetch-depth: 0  # For lastUpdated | 用于 lastUpdated

      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '20'

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

### Manual Trigger | 手动触发

1. Visit **Actions** tab | 访问 **Actions** 标签
2. Click "Deploy VitePress Docs to GitHub Pages" | 点击 "Deploy VitePress Docs to GitHub Pages"
3. Click **Run workflow** button | 点击 **Run workflow** 按钮
4. Select `main` branch | 选择 `main` 分支
5. Click green **Run workflow** | 点击绿色 **Run workflow**

---

## 🐛 Troubleshooting | 故障排查

### Problem: Build Fails | 问题：构建失败

**Check | 检查**:
```bash
npm run docs:build
# View error messages | 查看错误信息
```

**Common causes | 常见原因**:
- Markdown syntax errors | Markdown 语法错误
- Broken internal links | 内部链接失效
- Config file syntax errors | 配置文件语法错误

### Problem: 404 Page | 问题：404 页面

**Check base config | 检查 base 配置**:
```ts
// website/.vitepress/config.mjs
base: '/agno-Go/',  // Must be correct! | 必须正确！
```

### Problem: Missing Styles | 问题：样式丢失

**Clear cache | 清理缓存**:
```bash
rm -rf website/.vitepress/cache
rm -rf website/.vitepress/dist
npm run docs:build
```

### Problem: npm Permission Errors | 问题：npm 权限错误

```bash
# Fix npm cache ownership | 修复 npm 缓存权限
sudo chown -R $(whoami) ~/.npm

# Clean and reinstall | 清理并重新安装
npm cache clean --force
npm install
```

### Problem: Port Already in Use | 问题：端口已被占用

```bash
# Use different port | 使用不同端口
npm run docs:dev -- --port 8080
```

### Problem: GitHub Pages Not Updating | 问题：GitHub Pages 未更新

1. Check **Actions** tab for errors | 检查 **Actions** 标签是否有错误
2. Ensure **Pages** source is "GitHub Actions" | 确保 **Pages** 源是 "GitHub Actions"
3. Wait 2-3 minutes for CDN cache | 等待 2-3 分钟 CDN 缓存
4. Hard refresh browser (Ctrl+Shift+R) | 强制刷新浏览器

---

## 🔄 Update Workflow | 更新流程

### Daily Documentation Updates | 日常文档更新

```bash
# 1. Edit documentation | 编辑文档
nano website/guide/my-page.md

# 2. Local preview (optional) | 本地预览（可选）
npm run docs:dev

# 3. Commit and push | 提交并推送
git add website/guide/my-page.md
git commit -m "docs: update my-page guide"
git push origin main

# 4. Wait for auto-deployment (2-3 minutes) | 等待自动部署（2-3分钟）
# 5. Visit https://rexleimo.github.io/agno-Go/
```

---

## 🌐 Custom Domain (Optional) | 自定义域名（可选）

If you have a custom domain like `docs.agno-go.com`:

如果您有自定义域名如 `docs.agno-go.com`：

### Step 1: Add CNAME Record | 添加 CNAME 记录

In your DNS provider | 在 DNS 提供商添加:

```
Type: CNAME
Name: docs
Value: rexleimo.github.io
```

### Step 2: Configure GitHub Pages | 配置 GitHub Pages

1. In **Settings → Pages**
2. Enter **Custom domain**: `docs.agno-go.com`
3. Check **Enforce HTTPS** | 勾选 **强制 HTTPS**

### Step 3: Update VitePress Config | 更新 VitePress 配置

Edit | 编辑 `website/.vitepress/config.mjs`（或 `config.ts`）:

```ts
export default defineConfig({
  base: '/',  // Change to root | 改为根路径
  // ...
})
```

### Step 4: Add CNAME File | 添加 CNAME 文件

Create | 创建 `website/public/CNAME`:

```
docs.agno-go.com
```

---

## 📚 Resources | 资源

### i18n (Locales) | 多语言（Locales）
- Configure locales under `locales` in `config.mjs` with `root` and language keys (e.g., `zh`, `ja`, `ko`).
- Provide per-locale `themeConfig` for nav and sidebar (already present in this repo).
- Create matching content folders: `website/<lang>/` with `index.md`, `guide/`, `api/`, etc.

### Official Documentation | 官方文档
- [VitePress Docs](https://vitepress.dev/)
- [GitHub Pages](https://docs.github.com/en/pages)
- [GitHub Actions](https://docs.github.com/en/actions)
- [Markdown Guide](https://www.markdownguide.org/)
- [Vue 3](https://vuejs.org/)

### Example Websites | 示例网站
- [VitePress Official](https://vitepress.dev/)
- [Vue 3 Docs](https://vuejs.org/)
- [Vite Docs](https://vitejs.dev/)

---

## ✅ Deployment Checklist | 部署检查清单

Before pushing to production | 推送到生产环境前检查:

- [ ] All pages have content | 所有页面有内容
- [ ] Internal links work | 内部链接正常
- [ ] Code examples tested | 代码示例已测试
- [ ] Sidebar navigation complete | 侧边栏导航完整
- [ ] Base URL correct | Base URL 正确
- [ ] Search works | 搜索功能正常
- [ ] Mobile layout good | 移动布局良好
- [ ] Dark mode works | 暗色模式正常

---

## 📊 Monitoring | 监控

### View Deployment History | 查看部署历史

1. Visit **Actions** tab | 访问 **Actions** 标签
2. Click any workflow run | 点击任意工作流运行
3. View detailed logs and errors | 查看详细日志和错误

### Status Badge | 状态徽章

Add to README.md | 添加到 README.md:

```markdown
[![Docs Deploy](https://github.com/rexleimo/agno-Go/actions/workflows/deploy-docs.yml/badge.svg)](https://github.com/rexleimo/agno-Go/actions/workflows/deploy-docs.yml)
```

---

## 💡 Tips | 提示

### Best Practices | 最佳实践

- **Preview before pushing** | 推送前预览: Always test with `npm run docs:build && npm run docs:preview`
- **Keep it simple** | 保持简单: VitePress is fast because it's simple
- **Use Markdown** | 使用 Markdown: No need for custom components for most content
- **Link internally** | 内部链接: Use `/guide/` not full URLs
- **Commit often** | 经常提交: Small commits are easier to debug

### Performance | 性能

- GitHub Pages CDN has 1-3 minute cache delay | GitHub Pages CDN 有 1-3 分钟缓存延迟
- Use `Ctrl+Shift+R` to force browser cache refresh | 使用 `Ctrl+Shift+R` 强制刷新浏览器缓存
- VitePress build time: ~10-30 seconds | VitePress 构建时间：~10-30 秒

---

## 🎯 Next Steps | 后续步骤

### Complete Missing Pages | 完善缺失页面

The following pages need content | 以下页面需要添加内容:

**Guide Section | 指南部分**
- `website/guide/team.md` - Team guide | 团队指南
- `website/guide/workflow.md` - Workflow guide | 工作流指南
- `website/guide/models.md` - Models guide | 模型指南
- `website/guide/tools.md` - Tools guide | 工具指南
- `website/guide/memory.md` - Memory guide | 记忆指南

**API Section | API 部分**
- `website/api/agent.md` - Agent detailed API | Agent 详细 API
- `website/api/team.md` - Team detailed API | Team 详细 API
- `website/api/workflow.md` - Workflow detailed API | Workflow 详细 API
- `website/api/models.md` - Models detailed API | Models 详细 API
- `website/api/tools.md` - Tools detailed API | Tools 详细 API

**Examples Section | 示例部分**
- `website/examples/simple-agent.md`
- `website/examples/team-demo.md`
- `website/examples/rag-demo.md`

### Customize Theme | 自定义主题

1. Create `website/.vitepress/theme/index.ts`
2. Add custom styles in `website/.vitepress/theme/custom.css`
3. See [VitePress Theme Guide](https://vitepress.dev/guide/custom-theme)

---

## 🆘 Getting Help | 获取帮助

If you encounter issues | 如果遇到问题:

1. **Check Actions logs** | 查看 Actions 日志: Click failed workflow for detailed errors
2. **Check documentation** | 查看文档: [VitePress Deploy Guide](https://vitepress.dev/guide/deploy)
3. **GitHub Issues** | GitHub 问题: [Create Issue](https://github.com/rexleimo/agno-Go/issues)
4. **Community** | 社区: [VitePress Discord](https://chat.vitejs.dev/)

---

## 🎉 Success! | 成功！

Your documentation is ready! After pushing to GitHub:

您的文档已就绪！推送到 GitHub 后：

1. Wait 2-3 minutes | 等待 2-3 分钟
2. Visit | 访问: **https://rexleimo.github.io/agno-Go/**
3. Share with users! | 分享给用户！

---

**Happy documenting!** 🚀

**祝文档编写愉快！** 🚀

---

*Documentation generated with Claude Code @ 2025-10-04*
*文档由 Claude Code 生成 @ 2025-10-04*
