# Snow

静态站点生成器

## 核心特性

- **多格式** — Markdown、Org-mode、HTML
- **多语言** — 按目录、文件后缀或 FrontMatter 区分语言，内置 i18n
- **分类系统** — 自动生成标签/分类/作者等分类页面，支持时间归档
- **输出格式** — RSS、Atom、JSON 等自定义格式
- **模板** — 使用 Pongo2 模版，Django/Jinja2 风格语法
- **主题系统** — 可复用主题，站点目录优先覆盖
- **实时预览** — 内置开发服务器 + WebSocket livereload

## 快速开始

- 安装
  ```bash
  $ go install github.com/honmaple/snow@latest
  ```
  or
  ```bash
  $ brew install honmaple/tap/snow
  ```

- 创建站点
  ```bash
  $ snow init myblog
  $ cd myblog
  ```

- 新建文章
  ```bash
  $ mkdir -p content/posts
  $ cat > content/posts/hello.md << 'EOF'
  ---
  title: "Hello World"
  date: 2024-01-15
  tags: [intro]
  ---
  ## Hello World
  
  这是我的第一篇文章！
  EOF
  ```

- 启动开发服务器
  ```bash
  $ snow server -D
  INFO Copying static...
  INFO Done: in 6.084µs
  INFO Building en site...
  DEBU write page [posts/first-page.md] -> /posts/first-page/
  DEBU write page [posts/hello.md] -> /posts/hello/
  INFO Done: 0 sections, 2 pages, 0 hidden pages and 0 taxonomies in 5.59625ms
  INFO Listen http://127.0.0.1:8000 ...
  ```

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

完整文档参照 [docs/](docs/) 目录，或在线访问 [https://docs.honmaple.com/snow/](https://docs.honmaple.com/snow)。

## License

MIT
