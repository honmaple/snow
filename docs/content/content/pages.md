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

**Page Bundle**：包含 `index.{md,org,html}` 的目录视为一个页面整体。目录内其他文件可以作为附属资源，详见 [附件资源](/content/assets)。

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
| `assets`                  | []string | Page Bundle 附属资源白名单            |
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

## 模板查找

渲染 Page 时模板按顺序查找：

1. FrontMatter 中的 `template` 值
2. `page.html`

## 模板变量

| 属性 | 类型 | 说明 |
|------|------|------|
| `page.File` | `File` | 源文件信息 |
| `page.File.Path` | string | 相对内容根的文件路径 |
| `page.File.Dir` | string | 文件所在目录 |
| `page.File.Name` | string | 文件名 |
| `page.File.BaseName` | string | 不含扩展名的文件名 |
| `page.File.LanguageName` | string | 文件名后缀识别出的语言名 |
| `page.File.Ext` | string | 文件扩展名 |
| `page.FrontMatter` | FrontMatter | FrontMatter 数据 |
| `page.FrontMatter.{xxx}` | any | 自定义 FrontMatter 字段值 |
| `page.Toc` | []Heading | 内容目录 |
| `page.Lang` | string | 语言代码 |
| `page.Slug` | string | URL slug |
| `page.Title` | string | 页面标题 |
| `page.Description` | string | 页面描述 |
| `page.Summary` | string | 摘要 |
| `page.Content` | string | 渲染后 HTML |
| `page.RawContent` | string | 原始内容 |
| `page.Draft` | bool | 是否为草稿 |
| `page.Hidden` | bool | 是否隐藏 |
| `page.IsBundle` | bool | 是否为 Page Bundle |
| `page.WordCount` | int64 | 词数统计 |
| `page.ReadingTime` | int64 | 预计阅读时间（分钟） |
| `page.Date` | time.Time | 创建时间 |
| `page.Modified` | time.Time | 修改时间 |
| `page.Path` | string | 相对 URL |
| `page.Permalink` | string | 绝对 URL |
| `page.Section` | Section | 所属栏目 |
| `page.Assets` | Assets | Page Bundle 附件资源 |
| `page.Formats` | Formats | 其他输出格式 |
| `page.Ancestors()` | Sections | 从所属栏目开始向上的栏目列表，不包含页面自身 |

常用关联对象字段：

| 属性 | 说明 |
|------|------|
| `page.Section.Title` | 所属栏目标题 |
| `page.Assets[n].Path` | 附件输出相对 URL |
| `page.Assets[n].Permalink` | 附件输出绝对 URL |
| `page.Formats.Find("rss")` | 查找指定名称的输出格式 |

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
