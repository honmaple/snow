# Snow

静态站点生成器，使用 Go 语言编写。

## 快速开始

```bash
# 安装
go install github.com/honmaple/snow@latest

# 创建站点
snow init myblog
cd myblog

# 新建文章
mkdir -p content/posts
cat > content/posts/hello.md << 'EOF'
---
title: "Hello World"
date: 2024-01-15
tags: [intro]
---
## Hello World

这是我的第一篇文章！
EOF

# 启动开发服务器
snow server -D
# → http://127.0.0.1:8000
```

## CLI

```
snow init [dir]       创建新站点
snow build             构建
snow server            开发服务器（支持热重载）
snow hooks             查看已启用的插件
```

构建选项：

```
snow build --clean -C         清理输出
snow build --debug -D         调试模式
snow build --output-dir -o    指定输出目录
snow build --mode -m          指定构建模式（如 publish）
snow build --include-drafts   包含草稿
snow build --dry-run          预演（不写文件）
```

## 核心特性

- **多格式** — Markdown (goldmark)、Org-mode、HTML
- **多语言** — 按目录、文件后缀或 FrontMatter 区分语言，内置 i18n
- **Taxonomy** — 自动生成标签/分类/作者等分类页面，支持时间归档
- **分页** — Section 和 Taxonomy 级别分页
- **输出格式** — RSS、Atom、JSON 等自定义格式
- **Pongo2 模板** — Django/Jinja2 风格语法
- **主题系统** — 可复用主题，站点目录优先覆盖
- **实时预览** — 内置开发服务器 + WebSocket livereload

## 内容管理

```yaml
# FrontMatter
---
title: "文章标题"
date: 2024-01-15
tags: [go, web]
categories: [Programming/Go]
draft: false
---
```

支持 Markdown、Org-mode、HTML 三种格式。以 `_index.*` 命名的文件将目录标记为 Section（栏目）。Taxonomy 系统自动从 FrontMatter 字段生成分类页面。

## 配置

`config.yaml`（YAML 格式）：

```yaml
base_url: "http://127.0.0.1:8000"
title: "My Blog"
language: "en"

theme: "snow"            # 主题名，对应 themes/ 下目录
content_dir: "content"
output_dir: "output"

sections:
  posts:
    sort_by: "date desc"
    paginate: 10

taxonomies:
  tags:
  categories:

params:
  menus:                 # 导航栏链接
    - name: "About"
      url: "/pages/about/"
```

## 插件系统 (Hooks)

插件通过 Hooks 机制在内容处理流程中注入逻辑。默认启用 3 个，共 7 个内置插件。

```bash
$ snow hooks
assets(enabled), encrypt(enabled), shortcode(enabled), filter, minify, pelican, rewrite
```

## 模板

基于 [Pongo2](https://github.com/flosch/pongo2)，语法兼容 Django/Jinja2：

```html
<h1>{{ page.Title }}</h1>
<time>{{ page.Date | date:"2006-01-02" }}</time>
<div>{{ page.Content | safe }}</div>

{% for page in pages %}
  <li><a href="{{ page.Path }}">{{ page.Title }}</a></li>
{% endfor %}
```

## 文档

完整文档参照 [docs/](docs/) 目录，或在线访问 [https://snow.honmaple.com](https://snow.honmaple.me)。

## License

MIT
