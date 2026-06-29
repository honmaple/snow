---
title: "配置"
weight: 30
---

Snow 使用 YAML 格式配置，默认为站点根目录下的 `config.yaml`，可通过 `--config` 参数指定其他文件。

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
| `content_truncate_len` | int | `49` | 摘要截取长度 |
| `content_truncate_ellipsis` | string | `...` | 摘要后缀 |
| `ignored_content` | []string | — | 忽略内容 glob |
| `ignored_static` | []string | — | 忽略静态文件 glob |

`ignored_content` 与 `ignored_static` 按相对路径匹配。以 `_` 或 `.` 开头的内容文件默认忽略，`_index.{md,org,html}` 除外。

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
    prevent_pre_code: true
  markdown:
    enabled: true
  orgmode:
    enabled: true
  niklasfasching:
    enabled: false
  html:
    enabled: true
```

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `enabled` | bool | Markdown/Org-mode 为 `true`，HTML/niklasfasching 为 `false` | 是否启用该解析器 |
| `style` | string | `monokai` | chroma 语法高亮样式 |
| `show_toc` | bool | `true` | 显示文章目录 |
| `show_line_numbers` | bool | `true` | 显示行号 |
| `prevent_pre_code` | bool | `true` | 高亮代码块时避免额外包裹 pre/code |

常见样式：`monokai`、`github`、`dracula`、`solarized-dark`。

`markups._default` 为所有解析器提供默认渲染选项；`markups.markdown`、`markups.orgmode`、`markups.niklasfasching`、`markups.html` 可以分别覆盖。默认启用 Markdown 和 Org-mode，HTML 与 `niklasfasching` parser 已内置注册但默认未启用；解析 `.html` 内容文件时需要设置 `markups.html.enabled: true`。

`show_toc` 对 Markdown、Org-mode、`niklasfasching` 和 HTML 都生效。HTML parser 会根据 `h1`-`h6` 生成目录，并给缺少 `id` 的标题自动补充 `heading-...`。

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
    paginate: 0
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
| `path_style` | - | 输出路径后处理，见 [Path Style 配置](#path-style-配置) |
| `sort_by` | `date desc` | 页面排序 |
| `paginate` | `0` | 分页数，`0` 不分页 |
| `paginate_path` | 自动 | 分页路径 |
| `template` | — | 无默认，按 `section.html` 查找 |

未设置或设为空字符串时，分页器会按输出路径类型选择默认值：pretty 路径使用 `page/{number}/`，ugly 路径使用 `{name}{number}{extension}`；详见 [分页](/content/pagination)。

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
| `path_style` | - | 输出路径后处理，见 [Path Style 配置](#path-style-配置) |
| `template` | — | 无默认，按 `page.html` 查找 |
| `draft` | `false` | 标记为草稿 |
| `hidden` | `false` | 隐藏页面 |
| `lang` | 站点配置 | 语言 |

## Path Style 配置

`path_style` 用于在路径变量解析后做统一后处理，可用于 Page、Section、Taxonomy 和 Taxonomy Term。可选值为 `none`、`lower`、`slug`、`slug_unicode`，也可以用逗号按顺序组合，例如 `lower,slug`。

| 值 | 说明 |
|----|------|
| `none` 或空字符串 | 保持解析后的路径不变 |
| `lower` | 将输出路径全部转为小写 |
| `slug` | 按路径段 slug 化，并保留最后一个文件扩展名 |
| `slug_unicode` | 按路径段 slug 化并保留 Unicode 字符，不受 `slugify.preserve_unicode` 配置影响 |

## Slugify 配置

```yaml
slugify:
  lowercase: true
  preserve_unicode: false
  preserve_chars: ""
```

Slugify 配置用于 `{slug}`、`{path:slug}`、`{title:slug}`、`{term:slug}` 和 `path_style: slug`。默认会将 Unicode 字符转写为 ASCII，例如中文标题会生成类似 `ni-hao-world` 的 slug；如果希望保留中文等 Unicode 字符，可以设置 `slugify.preserve_unicode: true`。

| 关键配置项 | 默认值 | 说明 |
|------------|--------|------|
| `lowercase` | `true` | 是否转为小写 |
| `preserve_chars` | 空字符串 | 额外允许保留的字符，例如 `+` 或 `.` |
| `preserve_unicode` | `false` | 是否保留 Unicode 字母和数字，例如中文；默认会转写为 ASCII |

## Taxonomy 配置

```yaml
taxonomies:
  _default:
    path: "{taxonomy}/"
    sort_by: "name"
    term:
      path: "{taxonomy}/{term:slug}/"
      sort_by: "date desc"
  tags:
  categories:
  authors:
```

只需在 `taxonomies` 下列出名称即可启用。配置查找：`taxonomies.{name}.{key}` → `taxonomies._default.{key}`。

## Hook 配置

默认启用 `assets`、`encrypt`、`links`、`shortcode`。其他内置 Hook 只设置默认权重，不会自动挂载：

| Hook | 默认启用 | 默认权重 |
|------|----------|----------|
| `snakecase` | false | `10` |
| `assets` | true | `20` |
| `pelican` | false | `30` |
| `rewrite` | false | `30` |
| `filter` | false | `40` |
| `encrypt` | true | `50` |
| `links` | true | `55` |
| `shortcode` | true | `60` |
| `minify` | false | `70` |
| `alias` | false | `80` |

`weight` 越小越先执行；权重相同时按 Hook 名称排序。如果显式配置 `hooks.<name>.enabled: true`，但对应 Hook 未注册，构建会返回错误。

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

## 主题

```yaml
theme: "snow"
theme_dir: "themes"
```

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| `theme` | — | 主题名称，对应 `themes/` 下的目录名 |
| `theme_dir` | `themes` | 主题存放目录 |
