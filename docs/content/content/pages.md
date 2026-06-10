---
title: "页面 (Page)"
weight: 10
---

Page 是 Snow 中最基本的内容单元。

## 创建页面

在 `content/` 下创建 `.md`、`.org` 或 `.html` 文件：

```
content/
├── about.md              # 普通 Page → /about/
└── posts/
    ├── hello.md           # → /posts/hello/
    └── bundle/            # Page Bundle
        ├── index.md       # → /posts/bundle/
        └── image.png      # 附属资源
```

**Page Bundle**：包含 `index.{md,org,html}` 的目录视为一个页面整体，目录内其他文件作为附属资源。

## FrontMatter

```yaml
---
title: "文章标题"
slug: "custom-slug"
date: 2024-01-15 20:35:00
modified: 2024-02-01 10:00:00

draft: false
hidden: false
render: true

path: "custom/url/"
template: "custom-post.html"
aliases:
  - "/old-url/"

tags:
  - go
  - web
categories:
  - Programming/Go
---
```

### 所有 FrontMatter 字段

| 字段                       | 类型      | 说明                                |
|---------------------------|----------|------------------------------------|
| `title`                   | string   | 页面标题                             |
| `slug`                    | string   | URL slug，默认从标题生成              |
| `date`                    | datetime | 创建时间                             |
| `modified`                | datetime | 修改时间                             |
| `draft`                   | bool     | 草稿，构建时默认跳过                  |
| `hidden`                  | bool     | 隐藏页面，不出现在列表中               |
| `render`                  | bool     | 是否渲染                             |
| `path`                    | string   | 自定义输出路径                       |
| `template`                | string   | 自定义模板                           |
| `aliases`                 | []string | 重定向别名                           |
| `asset_path`              | string   | 附属资源输出路径                      |
| `formats.{name}.path`     | string   | 格式输出路径                         |
| `formats.{name}.template` | string   | 格式输出模板                         |
| `{taxonomy}`              | any      | 自定义分类字段（如 `tags`、`authors`） |

## 页面配置

```yaml
pages:
  _default:
    path: "{path:slug}/{slug}/"
  posts:
    path: "articles/{date:%Y}/{date:%m}/{slug}.html"
  pages:
    hidden: true
    template: "page.html"
  en:
    lang: "en"
  drafts:
    draft: true
```

配置级联查找：`pages.{目录路径}` → 父目录 → `pages._default`。

## 路径变量

| 变量               | 说明                | 示例           |
|-------------------|--------------------|---------------|
| `{date:%Y}`       | 年 (4位)            | `2024`        |
| `{date:%m}`       | 月 (2位)            | `01`          |
| `{date:%d}`       | 日 (2位)            | `15`          |
| `{date:%H}`       | 时 (2位)            | `20`          |
| `{lang}`          | 语言代码             | `en`          |
| `{lang:optional}` | 语言代码，默认语言为空 | `en` 或空      |
| `{path}`          | 文件位置路径         | `posts`       |
| `{path:slug}`     | 路径 slug 化        | `posts`       |
| `{slug}`          | 页面 slug           | `hello-world` |
| `{title}`         | 页面标题             | `Hello World` |

> 注意：代码中不存在 `{filename}` 变量。

## 模板查找

渲染 Page 时模板按顺序查找：

1. FrontMatter 中的 `template` 值
2. `page.html`

## 模板变量

| 属性                      | 说明            |
|--------------------------|----------------|
| `page.Title`             | 页面标题         |
| `page.Slug`              | URL slug       |
| `page.Lang`              | 语言代码         |
| `page.Date`              | 创建时间         |
| `page.Modified`          | 修改时间         |
| `page.Path`              | 相对 URL        |
| `page.Permalink`         | 绝对 URL        |
| `page.Summary`           | 摘要            |
| `page.Content`           | 渲染后 HTML     |
| `page.RawContent`        | 原始内容         |
| `page.FrontMatter.{xxx}` | 自定义字段值     |
| `page.Aliases`           | 重定向别名       |
| `page.Formats`           | 其他输出格式     |
| `page.Draft`             | 是否为草稿       |
| `page.Hidden`            | 是否隐藏         |

## 草稿 (Draft)

```yaml
---
draft: true
---
```

或通过配置批量标记：

```yaml
pages:
  drafts:
    draft: true
```

构建时默认忽略草稿，`--include-drafts` 可包含。
