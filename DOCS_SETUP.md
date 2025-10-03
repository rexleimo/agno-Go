# VitePress Documentation Setup Guide

This guide explains how to use and deploy the Agno-Go documentation website.

## 🎉 What's Been Set Up

A complete VitePress documentation website with:

- ✅ **Homepage** - Hero section with features showcase
- ✅ **Guide** - Quick start, installation, core concepts
- ✅ **API Reference** - Complete API documentation structure
- ✅ **Advanced Topics** - Architecture, performance, deployment
- ✅ **Examples** - Working code examples
- ✅ **GitHub Actions** - Auto-deploy to GitHub Pages
- ✅ **Search** - Built-in local search
- ✅ **Dark Mode** - Toggle dark/light theme
- ✅ **Mobile Responsive** - Works on all devices

## 📁 Project Structure

```
agno-Go/
├── package.json              # Node.js dependencies
├── website/                  # VitePress documentation
│   ├── .vitepress/
│   │   └── config.ts        # Site configuration
│   ├── index.md             # Homepage
│   ├── guide/               # User guides
│   │   ├── index.md         # What is Agno-Go?
│   │   ├── quick-start.md   # 5-minute tutorial
│   │   ├── installation.md  # Setup guide
│   │   └── agent.md         # Agent guide
│   ├── api/                 # API reference
│   │   └── index.md         # API overview
│   ├── advanced/            # Advanced topics
│   │   ├── architecture.md  # Architecture
│   │   ├── performance.md   # Benchmarks
│   │   └── deployment.md    # Deployment
│   └── examples/            # Examples
│       └── index.md         # Examples overview
└── .github/workflows/
    └── deploy-docs.yml      # Auto-deployment
```

## 🚀 Quick Start

### Prerequisites

- Node.js 18+ ([Download](https://nodejs.org/))
- npm (comes with Node.js)

### Local Development

```bash
# 1. Install dependencies
npm install

# 2. Start dev server
npm run docs:dev
```

The site will be available at **http://localhost:5173** with hot reload.

**注意**: 如果遇到 npm 权限错误:
```bash
# 修复 npm 缓存权限
sudo chown -R $(whoami) ~/.npm

# 清理并重新安装
npm cache clean --force
npm install
```

### Build for Production

```bash
# Build site
npm run docs:build

# Preview production build
npm run docs:preview
```

## 🌐 GitHub Pages Deployment

### Automatic Deployment (Recommended)

The documentation automatically deploys to GitHub Pages when you push to `main`:

1. **Push changes to GitHub**:
   ```bash
   git add .
   git commit -m "feat(docs): add VitePress documentation"
   git push origin main
   ```

2. **GitHub Actions** builds and deploys automatically

3. **Access your site** at:
   ```
   https://rexleimo.github.io/agno-Go/
   ```

### First-Time Setup

If GitHub Pages isn't enabled yet:

1. Go to your GitHub repository
2. Click **Settings** → **Pages**
3. Under **Source**, select:
   - **Source**: GitHub Actions ✅
4. Save

That's it! The next push will trigger deployment.

### Verify Deployment

```bash
# Wait 2-3 minutes for deployment
curl https://rexleimo.github.io/agno-Go/

# You should see the HTML homepage
```

## 📝 Editing Documentation

### Add New Pages

1. **Create markdown file** in appropriate directory:
   ```bash
   # Example: Add a new guide
   touch website/guide/my-new-guide.md
   ```

2. **Edit `website/.vitepress/config.ts`** to add to sidebar:
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

3. **Test locally**:
   ```bash
   npm run docs:dev
   ```

### Markdown Features

VitePress supports enhanced Markdown:

#### Code Blocks

\`\`\`go
package main

func main() {
    fmt.Println("Hello!")
}
\`\`\`

#### Custom Containers

::: tip
This is a tip
:::

::: warning
This is a warning
:::

::: danger STOP
Danger zone!
:::

#### Line Highlighting

\`\`\`go{2,4-6}
package main

import "fmt" // highlighted

func main() { // highlighted lines
    fmt.Println("Hello")
}
\`\`\`

## 🔧 Configuration

### Site Settings

Edit `website/.vitepress/config.ts`:

```ts
export default defineConfig({
  title: "Agno-Go",                    // Site title
  description: "Your description",     // Meta description
  base: '/agno-Go/',                   // Base URL (IMPORTANT!)

  themeConfig: {
    nav: [...],        // Top navigation
    sidebar: {...},    // Sidebar menu
    socialLinks: [...] // GitHub, etc.
  }
})
```

### Base URL

**IMPORTANT**: The `base` must match your repository name:

```ts
base: '/agno-Go/',  // For https://rexleimo.github.io/agno-Go/
```

If deploying to a custom domain:

```ts
base: '/',  // For https://yourdomain.com/
```

## 🐛 Troubleshooting

### npm Permission Errors

```bash
# Fix npm cache ownership
sudo chown -R $(whoami) ~/.npm

# Clean and reinstall
npm cache clean --force
npm install
```

### Port Already in Use

```bash
# Use different port
npm run docs:dev -- --port 8080
```

### Build Errors

```bash
# Clear cache and rebuild
rm -rf website/.vitepress/cache
rm -rf website/.vitepress/dist
rm -rf node_modules
npm install
npm run docs:build
```

### GitHub Pages Not Updating

1. Check **Actions** tab in GitHub for errors
2. Ensure **Pages** is set to "GitHub Actions" source
3. Wait 2-3 minutes for CDN cache
4. Hard refresh browser (Ctrl+Shift+R)

### 404 Errors on GitHub Pages

Check `base` in config matches repository name:

```ts
// Correct for https://rexleimo.github.io/agno-Go/
base: '/agno-Go/'

// NOT '/' unless using custom domain
```

## 📚 Next Steps

### Complete Missing Pages

Some pages are placeholders. Complete them:

- `website/guide/team.md`
- `website/guide/workflow.md`
- `website/guide/models.md`
- `website/guide/tools.md`
- `website/guide/memory.md`
- `website/api/agent.md` (detailed API)
- `website/api/team.md`
- `website/api/workflow.md`
- etc.

### Add More Examples

Create example pages in `website/examples/`:

- `website/examples/simple-agent.md`
- `website/examples/team-demo.md`
- `website/examples/rag-demo.md`

### Customize Theme

1. Create `website/.vitepress/theme/index.ts`
2. Add custom styles in `website/.vitepress/theme/custom.css`
3. See [VitePress Theme Guide](https://vitepress.dev/guide/custom-theme)

## 📖 Resources

- **VitePress Docs**: https://vitepress.dev/
- **Markdown Guide**: https://www.markdownguide.org/
- **GitHub Pages**: https://docs.github.com/en/pages
- **Vue 3**: https://vuejs.org/ (for custom components)

## ✅ Checklist

Before pushing to production:

- [ ] All pages have content (no placeholders)
- [ ] Internal links work (`/guide/`, `/api/`, etc.)
- [ ] Code examples are tested
- [ ] Sidebar navigation is complete
- [ ] Base URL is correct
- [ ] Search works (automatically indexed)
- [ ] Mobile layout looks good
- [ ] Dark mode works

## 🎯 Success!

Your documentation is ready! After pushing to GitHub:

1. Wait 2-3 minutes
2. Visit: **https://rexleimo.github.io/agno-Go/**
3. Share with users!

## 💡 Tips

- **Preview before pushing**: Always test with `npm run docs:build && npm run docs:preview`
- **Keep it simple**: VitePress is fast because it's simple
- **Use Markdown**: No need for custom components for most content
- **Link internally**: Use `/guide/` not full URLs
- **Commit often**: Small commits are easier to debug

---

**Need help?** Check [VitePress Discord](https://discord.gg/vitepress) or [GitHub Discussions](https://github.com/vuejs/vitepress/discussions).

Happy documenting! 🚀
