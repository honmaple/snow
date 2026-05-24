---
title: "Snow"
template: "index.html"
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
