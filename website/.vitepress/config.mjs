import { defineConfig } from 'vitepress'

// https://vitepress.dev/reference/site-config
export default defineConfig({
  title: "Agno-Go",
  description: "High-performance multi-agent system framework built with Go",
  base: '/agno-Go/',
  ignoreDeadLinks: true,

  head: [
    ['link', { rel: 'icon', href: '/agno-Go/favicon.ico' }],
    ['meta', { name: 'theme-color', content: '#3eaf7c' }],
    ['meta', { name: 'og:type', content: 'website' }],
    ['meta', { name: 'og:site_name', content: 'Agno-Go' }],
  ],

  themeConfig: {
    logo: '/logo.svg',

    nav: [
      { text: 'Guide', link: '/guide/' },
      { text: 'API Reference', link: '/api/' },
      { text: 'Advanced', link: '/advanced/' },
      { text: 'Examples', link: '/examples/' },
      {
        text: 'v1.0.0',
        items: [
          { text: 'Changelog', link: 'https://github.com/rexleimo/agno-Go/blob/main/CHANGELOG.md' },
          { text: 'Release Notes', link: '/release-notes' }
        ]
      }
    ],

    sidebar: {
      '/guide/': [
        {
          text: 'Introduction',
          items: [
            { text: 'What is Agno-Go?', link: '/guide/' },
            { text: 'Quick Start', link: '/guide/quick-start' },
            { text: 'Installation', link: '/guide/installation' }
          ]
        },
        {
          text: 'Core Concepts',
          items: [
            { text: 'Agent', link: '/guide/agent' },
            { text: 'Team', link: '/guide/team' },
            { text: 'Workflow', link: '/guide/workflow' },
            { text: 'Models', link: '/guide/models' },
            { text: 'Tools', link: '/guide/tools' },
            { text: 'Memory', link: '/guide/memory' }
          ]
        }
      ],
      '/api/': [
        {
          text: 'API Reference',
          items: [
            { text: 'Overview', link: '/api/' },
            { text: 'Agent', link: '/api/agent' },
            { text: 'Team', link: '/api/team' },
            { text: 'Workflow', link: '/api/workflow' },
            { text: 'Models', link: '/api/models' },
            { text: 'Tools', link: '/api/tools' },
            { text: 'Memory', link: '/api/memory' },
            { text: 'Types', link: '/api/types' },
            { text: 'AgentOS Server', link: '/api/agentos' }
          ]
        }
      ],
      '/advanced/': [
        {
          text: 'Advanced Topics',
          items: [
            { text: 'Architecture', link: '/advanced/architecture' },
            { text: 'Performance', link: '/advanced/performance' },
            { text: 'Deployment', link: '/advanced/deployment' },
            { text: 'Testing', link: '/advanced/testing' }
          ]
        }
      ],
      '/examples/': [
        {
          text: 'Examples',
          items: [
            { text: 'Overview', link: '/examples/' },
            { text: 'Simple Agent', link: '/examples/simple-agent' },
            { text: 'Claude Agent', link: '/examples/claude-agent' },
            { text: 'Ollama Agent', link: '/examples/ollama-agent' },
            { text: 'Team Demo', link: '/examples/team-demo' },
            { text: 'Workflow Demo', link: '/examples/workflow-demo' },
            { text: 'RAG Demo', link: '/examples/rag-demo' }
          ]
        }
      ]
    },

    socialLinks: [
      { icon: 'github', link: 'https://github.com/rexleimo/agno-Go' }
    ],

    footer: {
      message: 'Released under the MIT License.',
      copyright: 'Copyright © 2025 Agno-Go Team'
    },

    search: {
      provider: 'local'
    },

    editLink: {
      pattern: 'https://github.com/rexleimo/agno-Go/edit/main/website/:path',
      text: 'Edit this page on GitHub'
    },

    lastUpdated: {
      text: 'Last updated',
      formatOptions: {
        dateStyle: 'medium',
        timeStyle: 'short'
      }
    }
  },

  markdown: {
    lineNumbers: true,
    theme: {
      light: 'github-light',
      dark: 'github-dark'
    }
  }
})
