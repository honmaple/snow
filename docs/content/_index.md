---
title: "Snow Docs"
sort_by: "weight"
template: "index.html"

params.hero:
  name: "红枫云盘"
  label: "多协议云盘文件管理"
  description: "A fast, elegant documentation theme for Zola, built with Tailwind CSS and designed for technical teams that care about clarity."
  actions:
    - text: "Get started"
      link: "/getting-started/"
      theme: "primary"
    - text: "Download"
      link: "https://github.com/honmaple/snow"
  preview:
      image: "/aaa.png"
      image_alt: "Preview screenshot"
      brand: "Snow Docs"
      title: "Ship beautiful docs with Zola"
      description: "Fast pages, clear structure, responsive navigation, and local Tailwind builds."
      code_label: "config.toml"
      code_lines: 
        - "title = \"Snow Docs\""
        - "build_search_index = true"
        - "template = \"landing.html\""
      sidebar:
        - label: "Advanced"
          active: false
        - label: "Guide"
          active: false
        - label: "Advanced"
          active: false
        - label: "Nested Pages"
          active: true
        - label: "Markdown Usage"
          active: false
      stats:
        - label: "Local CSS"
          value: "Tailwind 4"
        - label: "Docs shell"
          value: "Responsive"

params.features:
  - title: 开源
    description: 软件的完整源代码托管在Github，你可以自由查看并使用，无需担心软件被植入后门
    link: https://github.com/honmaple/maple-file
  - title: 跨平台
    description: 支持Web、Android、MacOS和Windows
  - title: 多协议
    description: 支持本地存储、S3、Webdav、FTP、SFTP、SMB、又拍云、Alist、Mirror、115、夸克网盘
  - title: 文件操作和预览
    description: 支持文件复制、移动、重命名、上传、下载，以及视频、音频、图片和文本文件的预览
  - title: 文件加密和压缩
    description: 保护隐私，通过AES加密技术避免你的文件被泄漏或被审查
  - title: 文件同步和备份
    description: 支持各存储之间的备份和同步
  - title: 回收站
    description: 文件误删除也能重新恢复
  - title: 多语言
    description: 支持中文、English
  - title: 多主题
    description: 支持多种颜色主题切换

params.intros:
  - title: "Built for documentation that gets used."
    description: "The homepage pulls positioning and feature content from Markdown front matter while the rest of the theme keeps the docs experience predictable."
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

[GitHub](https://github.com/honmaple/snow)