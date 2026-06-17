---
title: "栏目 (Section)"
weight: 20
---

Section 是组织 Pages 的树状层级结构。包含 `_index.{md,org,html}` 文件的目录即为 Section。

`.html` Section 文件需要先启用 `markups.html.enabled: true`；Markdown 和 Org-mode 默认启用。

## 目录示例

```
content/
├── _index.md              # 根 Section → /
└── posts/
    ├── _index.md          # Section → /posts/
    ├── article1.md        # → /posts/article1/
    └── tutorials/         # 子 Section → /posts/tutorials/
        ├── _index.md
        └── intro.md       # → /posts/tutorials/intro/
```

## Section FrontMatter

`_index.md` 中的 FrontMatter 控制 Section 行为：

```yaml
---
title: "文章列表"
sort_by: "date desc"
paginate: 10
template: "custom-section.html"
---
这是 Section 描述内容，在模板中通过 `section.Content` 呈现。
```

### FrontMatter 字段

| 字段 | 类型 | 说明 |
|------|------|------|
| `title` | string | 栏目标题 |
| `slug` | string | 栏目 slug |
| `sort_by` | string | 页面排序，格式 `"field direction"` |
| `path` | string | 自定义输出路径，为空跳过渲染 |
| `template` | string | 自定义模板 |
| `paginate` | int | 每页条目数，`0` 不分页 |
| `paginate_path` | string | 分页路径模板 |
| `paginate_filter_by` | string | 分页前过滤 |
| `render` | bool | 是否渲染，`false` 跳过 |
| `assets` | []string | 附属资源文件 |
| `formats.{name}.path` | string | 格式输出路径 |
| `formats.{name}.template` | string | 格式输出模板 |

注意：`title` 留空时自动取目录名，根 Section 默认为 `index`。`assets` 字段用于声明栏目附属资源，详见 [附件资源](/content/assets)。

## 配置

```yaml
sections:
  _default:
    path: "{path:slug}/"
    sort_by: "date desc"
    paginate: 0
    paginate_path: "{name}{number:optional}{extension}"
  posts:
    sort_by: "date desc"
    paginate: 5
    template: "custom.html"
  pages:
    path: ""
```

Section 类型由目录名确定（如 `posts`、`pages`、`tutorials`）。配置级联查找：

1. `sections.{目录路径}` — 如 `posts/tutorials`
2. 父目录 — 如 `posts`
3. `sections._default`

## 路径变量

`path` 支持的占位符：

| 变量 | 说明 | 示例 |
|------|------|------|
| `{lang}` | 语言代码 | `en` |
| `{lang:optional}` | 语言代码，默认语言时为空 | `en` 或空 |
| `{path}` | Section 目录路径 | `posts/tutorials` |
| `{path:slug}` | 路径 slug 化 | `posts-tutorials` |
| `{section}` | Section 标题 | `Tutorials` |
| `{section:slug}` | 标题 slug 化 | `tutorials` |

> `{section}` 取值为 `section.Title`（FrontMatter 中设置或自动生成），而非目录名。

## 排序

`sort_by` 支持多字段逗号分隔：

```yaml
sort_by: "date desc, title asc"
```

可用字段：`date`（创建时间）、`modified`（修改时间）、`title`、`weight`，以及任意 FrontMatter 字段。

## 模板查找

渲染 Section 时模板按顺序查找：

1. FrontMatter 中的 `template` 值
2. `section.html`
3. 根 Section (`_index.md` 位于 `content/` 根目录) 额外尝试 `index.html`

## 过滤表达式

```yaml
paginate_filter_by: "'emacs' in tags and not draft"
```

表达式中的字段为 Page 的 FrontMatter 值。

## 模板变量

| 属性 | 类型 | 说明 |
|------|------|------|
| `section.File` | `File` | 源文件信息 |
| `section.File.Path` | string | 相对内容根的 `_index` 文件路径 |
| `section.File.Dir` | string | Section 所在目录 |
| `section.File.Name` | string | 文件名 |
| `section.File.BaseName` | string | 不含扩展名的文件名 |
| `section.File.LanguageName` | string | 文件名后缀识别出的语言名 |
| `section.File.Ext` | string | 文件扩展名 |
| `section.FrontMatter` | FrontMatter | FrontMatter 数据 |
| `section.FrontMatter.{xxx}` | any | 自定义 FrontMatter 字段值 |
| `section.Toc` | []Heading | 内容目录 |
| `section.Lang` | string | 语言代码 |
| `section.Slug` | string | 栏目 slug |
| `section.Title` | string | 栏目标题 |
| `section.Description` | string | 栏目描述 |
| `section.Summary` | string | 摘要 |
| `section.Content` | string | `_index.md` 正文渲染内容 |
| `section.RawContent` | string | `_index.md` 原始内容 |
| `section.WordCount` | int64 | 词数统计 |
| `section.ReadingTime` | int64 | 预计阅读时间（分钟） |
| `section.Path` | string | 相对 URL |
| `section.Permalink` | string | 绝对 URL |
| `section.Pages` | Pages | 栏目下普通页面列表 |
| `section.HiddenPages` | Pages | 栏目下隐藏页面列表 |
| `section.Assets` | Assets | Section 附件资源 |
| `section.Formats` | Formats | 其他输出格式 |
| `section.Parent` | Section | 父栏目 |
| `section.Children` | Sections | 子栏目列表 |
| `section.IsHome()` | bool | 是否为首页 Section |
| `section.Ancestors()` | Sections | 从父栏目开始向上的栏目列表，不包含自身 |
| `section.RecursivePages()` | Pages | 当前栏目和子栏目下的普通页面 |
| `section.RecursiveHiddenPages()` | Pages | 当前栏目和子栏目下的隐藏页面 |
| `paginator` | Paginator | 分页对象（启用分页时） |

常用关联对象字段：

| 属性 | 说明 |
|------|------|
| `section.Parent.Title` | 父栏目标题 |
| `section.Children[n].Title` | 子栏目标题 |
| `section.Assets[n].Path` | 附件输出相对 URL |
| `section.Assets[n].Permalink` | 附件输出绝对 URL |
| `section.Formats.Find("rss")` | 查找指定名称的输出格式 |
