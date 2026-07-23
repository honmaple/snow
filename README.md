# Snow

A static site generator.

## Features

- **Multiple content formats** — Markdown, Org-mode, and HTML
- **Multilingual sites** — Detect languages by file suffix or Front Matter, with built-in i18n support
- **Taxonomies** — Automatically generate tag, category, author, and archive pages
- **Output formats** — Custom formats such as RSS, Atom, and JSON
- **Templates** — Pongo2 templates with Django/Jinja2-style syntax
- **Themes** — Reusable themes with site-level overrides
- **Live preview** — Built-in development server with WebSocket livereload

## Quick Start

- Install
  ```bash
  $ go install github.com/honmaple/snow@latest
  ```
  or
  ```bash
  $ brew install honmaple/tap/snow
  ```

- Create a site
  ```bash
  $ snow init myblog
  $ cd myblog
  ```

- Create a post
  ```bash
  $ mkdir -p content/posts
  $ cat > content/posts/hello.md << 'EOF'
  ---
  title: "Hello World"
  date: 2024-01-15
  tags: [intro]
  ---
  ## Hello World
  
  This is my first post!
  EOF
  ```

- Start the development server
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

## Content

```yaml
# FrontMatter
---
title: "Post Title"
date: 2024-01-15
tags: [go, web]
categories: [Programming/Go]
draft: false
---
```

Snow supports Markdown, Org-mode, and HTML. Files named `_index.*` mark a directory as a section. The taxonomy system automatically generates taxonomy pages from Front Matter fields.

## Configuration

`config.yaml` uses YAML:

```yaml
base_url: "http://127.0.0.1:8000"
title: "My Blog"
language: "en"

theme: "snow"            # Theme name, matching a directory under themes/
output_dir: "output"

sections:
  posts:
    sort_by: "date desc"
    paginate: 10

taxonomies:
  tags:
  categories:

params:
  menus:                 # Navigation links
    - name: "About"
      url: "/pages/about/"
```

## Hooks

Hooks let plugins inject behavior into the content processing pipeline. Snow includes built-in hooks for assets, redirects, content links, shortcodes, and more.

```bash
$ snow hooks
mount, snakecase, assets(enabled), pelican, rewrite, filter, encrypt(enabled), links(enabled), shortcode(enabled), minify, alias
```

## Templates

Snow uses [Pongo2](https://github.com/flosch/pongo2), with Django/Jinja2-compatible syntax:

```html
<h1>{{ page.Title }}</h1>
<time>{{ page.Date | date:"2006-01-02" }}</time>
<div>{{ page.Content | safe }}</div>

{% for page in pages %}
  <li><a href="{{ page.Path }}">{{ page.Title }}</a></li>
{% endfor %}
```

## Documentation

See the [docs/](docs/) directory for complete documentation, or visit [https://docs.honmaple.com/snow/](https://docs.honmaple.com/snow).

## License

MIT
