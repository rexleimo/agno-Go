# 🚀 VitePress 文档快速部署

## 一键部署（3 步）

### 1️⃣ 推送到 GitHub

```bash
git add .
git commit -m "feat(docs): add VitePress documentation website"
git push origin main
```

### 2️⃣ 启用 GitHub Pages

访问: https://github.com/rexleimo/agno-Go/settings/pages

- **Source** 选择: **GitHub Actions** ✅

### 3️⃣ 访问网站（2-3 分钟后）

🎉 **https://rexleimo.github.io/agno-Go/**

---

## 本地开发（可选）

```bash
# 安装依赖
npm install

# 启动开发服务器
npm run docs:dev
# 访问 http://localhost:5173
```

**权限错误修复**:
```bash
sudo chown -R $(whoami) ~/.npm
npm cache clean --force
npm install
```

---

## 文件结构

```
agno-Go/
├── package.json                      # ✅ 已创建
├── website/                          # ✅ 文档目录
│   ├── .vitepress/config.ts         # ✅ 配置文件
│   ├── index.md                     # ✅ 首页
│   ├── guide/                       # ✅ 指南
│   ├── api/                         # ✅ API
│   ├── advanced/                    # ✅ 高级
│   └── examples/                    # ✅ 示例
└── .github/workflows/deploy-docs.yml # ✅ 自动部署
```

---

## 故障排查

### ❌ 构建失败

```bash
# 查看 Actions 日志
# GitHub → Actions → 点击失败的运行 → 查看详情
```

### ❌ 页面 404

检查配置:
```ts
// website/.vitepress/config.ts
base: '/agno-Go/',  // 必须与仓库名匹配！
```

### ❌ 样式丢失

清理缓存:
```bash
rm -rf website/.vitepress/cache
rm -rf website/.vitepress/dist
npm run docs:build
```

---

## 详细文档

- 📖 [完整总结](VITEPRESS_SUMMARY.md) - 所有功能和文件清单
- 🔧 [本地开发](DOCS_SETUP.md) - VitePress 开发指南
- 🌐 [部署指南](GITHUB_PAGES_SETUP.md) - GitHub Pages 详细步骤

---

**就这么简单！** 🎉
