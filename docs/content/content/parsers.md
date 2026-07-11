---
title: "解析器 (Parser)"
weight: 5
---

Snow 通过内容解析器把 `.md`、`.org`、`.html` 文件转换为统一的 Page / Section 数据。

## 内置解析器

| 解析器 | 扩展名 | 引擎 | 默认启用 | 说明 |
|--------|--------|------|----------|------|
| `markdown` | `.md` | goldmark | 是 | Markdown 内容，支持 YAML/TOML FrontMatter |
| `orgmode` | `.org` | org-golang | 是 | 默认 Org-mode 解析器 |
| `niklasfasching` | `.org` | niklasfasching/go-org | 否 | 可选 Org-mode 解析器 |
| `html` | `.html` | Go HTML parser | 否 | HTML 文档或片段 |

HTML 和 `niklasfasching` parser 已在 CLI 中注册，但默认未启用。需要使用时在 `markups` 中开启：

```yaml
markups:
  html:
    enabled: true
  niklasfasching:
    enabled: true
```

`orgmode` 和 `niklasfasching` 都处理 `.org` 文件。如果同时启用，当前注册顺序下 `niklasfasching` 会优先接管 `.org`，默认的 `orgmode` 不再处理同一扩展名。

## 元数据

| 格式 | 元数据来源 |
|------|------------|
| Markdown | YAML `---`、TOML `+++`、或文件开头的 `key: value` 行 |
| Org-mode | `:PROPERTIES:` drawer、文件开头的 `#+KEY:`、`#+PROPERTY:` |
| HTML | `<head>` 中的 `<title>`、`<meta>`、`<link>`、`<script>` |

Markdown FrontMatter 示例：

```markdown
---
title: "我的文章"
date: 2024-01-15
tags:
  - go
  - web
draft: false
---
```

启用 `markups.markdown.directive_blocks` 后，Markdown 支持类似 Org-mode 的指令块：

```yaml
markups:
  markdown:
    directive_blocks: true
```

````markdown
:::export html
<div class="raw-html">
  <span>原样输出 HTML</span>
</div>
:::

:::center
这里的 **Markdown** 会继续解析，并包裹在 `<div style="text-align: center;">` 中。
:::

:::quote
这里的 **Markdown** 会继续解析，并包裹在 `<blockquote>` 中。
:::

:::shortcode notice type=info
这里的 **Markdown** 会继续解析，并作为 shortcode body 传入。

- static-site
- go
:::
````

`:::shortcode` 会渲染为 `<shortcode ...>...</shortcode>`，块内 Markdown 会继续解析，再作为 shortcode body 传入。若 body 中还有嵌套 shortcode，会先渲染最内层 shortcode，再把渲染后的 HTML 传给外层。开头行中第二个字段是 shortcode 名称，后续 `key=value` 会转为 shortcode 属性。需要在 body 中传递 YAML、TOML 等原始配置时，使用普通 HTML 标签形式 `<shortcode name>...</shortcode>`。

HTML parser 会把 `<head>` 中的标签转换为 FrontMatter：

| HTML 标签 | 写入字段 |
|-----------|----------|
| `<title>` | `title` |
| `<meta name="date" content="...">` | `date` |
| `<meta property="og:title" content="...">` | `og.title` |
| `<meta itemprop="description" content="...">` | `description` |
| `<link rel="stylesheet" href="...">` | 追加到 `links` |
| `<script src="...">` | 追加到 `scripts` |

## 正文与摘要

| 格式 | 正文来源 | 摘要分隔符 |
|------|----------|------------|
| Markdown | Markdown 渲染结果 | `<!--more-->` |
| Org-mode | Org 渲染结果 | `#+snow: more` 或 `#+html: <!--more-->` |
| HTML | `<body>` 子节点；没有完整文档结构时把 HTML 片段作为正文 | `<!--more-->` |

Org-mode 的摘要分隔符使用 Snow 专用 keyword：

```org
summary
#+snow: more
content
```

也可以使用导出 HTML keyword，便于沿用已有 Org 内容：

```org
summary
#+html: <!--more-->
content
```

## Toc

启用 `markups._default.show_toc` 或对应 parser 的 `show_toc` 后，解析器会生成 `Toc`。

Markdown 和 Org-mode 根据标题结构生成目录。HTML parser 会扫描 `h1`-`h6`，保留已有 `id`，并为缺少 `id` 的标题自动补上 `heading-...`。

## 模板 Filter

| Filter | 说明 |
|--------|------|
| `parser:"markdown"` | 使用 Markdown parser 把字符串转为 HTML |
| `parser:"orgmode"` | 使用默认 Org-mode parser 把字符串转为 HTML |
| `parser:"niklasfasching"` | 使用 niklasfasching/go-org 把字符串转为 HTML；只有 `markups.niklasfasching.enabled: true` 时可用 |

`parser` filter 只会调用当前配置中已启用的解析器。解析器未启用或格式名不存在时，模板渲染会返回错误。

## 错误处理与限制

Markdown 与 Org-mode 解析器支持最长约 1MB 的单行内容。Markdown FrontMatter 必须使用同类型 fence 闭合；Org-mode 的 `:PROPERTIES:` drawer 必须以 `:END:` 闭合，否则构建会报错。

HTML parser 基于 Go HTML tree parser。解析错误会直接返回并中止构建。
