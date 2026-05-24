# 配置

Snow 使用 YAML 格式配置，默认为站点根目录下的 `config.yaml`，可通过 `--config` 参数指定其他文件。

## 完整示例

```yaml
#─────────────────────────────────────
# 站点信息
#─────────────────────────────────────
base_url: "http://127.0.0.1:8000"
title: "My Blog"
description: "一个用 Snow 构建的博客"
author: "honmaple"
language: "en"

#─────────────────────────────────────
# 目录
#─────────────────────────────────────
content_dir: "content"
static_dir: "static"
output_dir: "output"
theme_dir: "themes"

#─────────────────────────────────────
# 内容处理
#─────────────────────────────────────
slugify: true
content_truncate_len: 49
content_truncate_ellipsis: "..."

# 忽略的内容/静态文件（目录以 / 结尾）
ignored_content:
  - "drafts/"
ignored_static:
  - "extra/"

#─────────────────────────────────────
# 主题
#─────────────────────────────────────
theme:
  name: "snow"
  config: "theme.yaml"
  override: "layouts"
```

## 站点信息

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `base_url` | string | `http://127.0.0.1:8000` | 站点根 URL |
| `title` | string | `snow` | 站点标题 |
| `description` | string | `snow is a static site generator.` | 站点描述 |
| `author` | string | `honmaple` | 站点作者 |
| `language` | string | `en` | 默认语言代码 |

## 目录配置

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| `content_dir` | `content` | 内容目录 |
| `static_dir` | `static` | 静态文件目录 |
| `output_dir` | `output` | 构建输出目录 |
| `theme_dir` | `themes` | 主题存放目录 |

## 内容处理

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `slugify` | bool | `true` | URL slug 转换 |
| `content_truncate_len` | int | `49` | 摘要截取长度 |
| `content_truncate_ellipsis` | string | `...` | 摘要后缀 |
| `ignored_content` | []string | — | 忽略内容 glob |
| `ignored_static` | []string | — | 忽略静态文件 glob |

## 多环境 (Modes)

```yaml
base_url: "http://127.0.0.1:8000"

modes:
  publish:
    base_url: "https://example.com"
  develop:
    debug: true
    include: "develop.yaml"
```

通过 `--mode publish` 构建时，`base_url` 等配置会被覆盖。`include` 可引用外部文件合并配置。

## 语法高亮 (Markups)

```yaml
markups:
  _default:
    style: "monokai"
    show_toc: true
    show_line_numbers: true
```

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `style` | string | `monokai` | chroma 语法高亮样式 |
| `show_toc` | bool | `true` | 显示文章目录 |
| `show_line_numbers` | bool | `true` | 显示行号 |

常见样式：`monokai`、`github`、`dracula`、`solarized-dark`。

## 输出格式 (Formats)

```yaml
formats:
  rss:
    template: "partials/rss.xml"
  atom:
    template: "partials/atom.xml"
```

每个内容类型可通过 `formats.{name}.path` 和 `formats.{name}.template` 覆盖默认值。

## Section 配置

```yaml
sections:
  _default:
    path: "{path:slug}/"
    sort_by: "date desc"
    paginate: 10
    paginate_path: "{name}{number:optional}{extension}"
  posts:
    sort_by: "date desc"
    paginate: 5
    template: "custom.html"
  pages:
    path: ""
```

配置查找：`sections.{目录名}` → 父目录 → `sections._default`。

| 关键配置项 | 默认值 | 说明 |
|------------|--------|------|
| `path` | `{path:slug}/` | 输出路径，为空禁用渲染 |
| `sort_by` | `date desc` | 页面排序 |
| `paginate` | `10` | 分页数，`0` 不分页 |
| `paginate_path` | `{name}{number:optional}{extension}` | 分页路径 |
| `template` | — | 无默认，按 `section.html` 查找 |

## Page 配置

```yaml
pages:
  _default:
    path: "{path:slug}/{slug}/"
  posts:
    path: "articles/{date:%Y}/{date:%m}/{slug}.html"
  pages:
    hidden: true
    template: "page.html"
  drafts:
    draft: true
```

配置查找：`pages.{目录名}` → 父目录 → `pages._default`。

| 关键配置项 | 默认值 | 说明 |
|------------|--------|------|
| `path` | `{path:slug}/{slug}/` | 输出路径 |
| `template` | — | 无默认，按 `page.html` 查找 |
| `draft` | `false` | 标记为草稿 |
| `hidden` | `false` | 隐藏页面 |
| `lang` | 站点配置 | 语言 |

## Taxonomy 配置

```yaml
taxonomies:
  _default:
    path: "{taxonomy}/"
    sort_by: "name"
    term:
      path: "{taxonomy}/{term:slug}/"
      sort_by: "date desc"
      paginate_path: "{name}{number:optional}{extension}"
  tags:
  categories:
  authors:
```

只需在 `taxonomies` 下列出名称即可启用。配置查找：`taxonomies.{name}.{key}` → `taxonomies._default.{key}`。

## 多语言

```yaml
language: "zh"

languages:
  en:
    content_dir: "content/en"
    translations: "i18n/en.yaml"
  fr:
    translations:
      - id: "tags"
        tr: "Tags"
```

每个语言可覆盖任何全局配置。

## 插件

```yaml
hooks:
  assets:
    enabled: true
  encrypt:
    enabled: true
    weight: 2
    option:
      password: "123456"
  shortcode:
    enabled: true
    weight: 1
  minify:
    enabled: false
    option:
      html: true
      css: true
      js: true
  rewrite:
    enabled: false
    option:
      - src: "tag"
        dst: "tags"
        type: "list"
```

`assets`、`encrypt`、`shortcode` 默认启用。`weight` 控制执行顺序，越大越先执行。

## 主题

```yaml
theme: "snow"
theme_dir: "themes"
```

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| `theme` | — | 主题名称，对应 `themes/` 下的目录名 |
| `theme_dir` | `themes` | 主题存放目录 |
