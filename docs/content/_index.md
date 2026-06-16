---
title: "Snow"
description: "一个简洁、快速、可扩展的静态站点生成器，支持 Markdown、Org-mode、HTML、多语言、Taxonomy、分页、模板和 Hook 插件流程。"
sort_by: "weight"
template: "index.html"

params.hero:
  name: "Snow"
  label: "静态站点生成器"
  actions:
    - text: "Get started"
      link: "/getting-started/"
      theme: "primary"
    - text: "GitHub"
      link: "https://github.com/honmaple/snow"
  preview:
      brand: "Snow"
      title: "Build static sites from plain content"
      description: "Single binary, Pongo2 templates, content parsing, assets, shortcodes and local preview in one workflow."
      code_label: "config.yaml"
      code_lines: 
        - "title: \"My Blog\""
        - "content_dir: \"content\""
        - "hooks:"
        - "  shortcode:"
        - "    enabled: true"
      sidebar:
        - label: "Getting Started"
          active: false
        - label: "Content"
          active: false
        - label: "Templates"
          active: false
        - label: "Hooks"
          active: true
        - label: "Configuration"
          active: false
      stats:
        - label: "Formats"
          value: "MD / Org / HTML"
        - label: "Runtime"
          value: "Single binary"

params.features:
  - title: 单二进制
    description: 使用 Go 编写，构建后就是一个 CLI 程序，适合本地预览、CI 构建和简单部署。
  - title: 多内容格式
    description: 支持 Markdown、Org-mode 和 HTML，内置 FrontMatter、摘要、目录和语法高亮处理。
  - title: Pongo2 模板
    description: 使用 Django/Jinja 风格模板语法，并提供页面、栏目、分类、i18n、数据加载等模板函数。
  - title: 内容组织
    description: 内置 Page、Section、Taxonomy、分页、多语言和自定义输出格式，覆盖博客和文档站常见结构。
  - title: Hook 插件
    description: 通过 assets、shortcode、encrypt、filter、rewrite、minify、snakecase 等 Hook 扩展构建流程。
  - title: 开发预览
    description: 提供本地开发服务器、文件监听和 WebSocket 热重载，便于边写内容边查看结果。

params.intros:
  - title: "为内容站点和技术文档准备的轻量工具。"
    description: "Snow 把内容解析、模板渲染、静态资源处理和插件扩展放在同一条清晰的构建链路里，保持配置简单，也保留足够的定制空间。"
---

Snow 是一个简洁的静态站点生成器，使用 Go 语言编写。

## 为什么选择 Snow？

- **轻量** — 单二进制文件，零运行时依赖
- **快速** — 亚秒级构建
- **多格式** — Markdown (goldmark)、Org-mode、HTML
- **多语言** — 内置 i18n
- **灵活** — 插件系统、Taxonomy、分页、自定义输出格式
- **实时预览** — 开发服务器 + WebSocket 热重载

## 快速开始

```bash
go install github.com/honmaple/snow@latest
snow init myblog && cd myblog

cat > content/posts/hello.md << 'EOF'
---
title: "Hello World"
date: 2024-01-15
tags: [intro]
---
## Hello World

我的第一篇文章！
EOF

snow server -D
# → http://127.0.0.1:8000
```

## 文档

- [快速开始](/getting-started) — 安装、构建、开发服务器
- [目录结构](/directory-structure) — 站点文件布局
- [配置](/configuration) — 完整配置参考
- [内容管理](/content) — 页面、栏目、分类、多语言
- [模板](/templates) — 模板语法与变量
- [主题](/themes) — 主题创建与使用
- [插件](/hooks) — 插件系统

