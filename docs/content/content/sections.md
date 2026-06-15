---
title: "栏目 (Section)"
weight: 20
---

Section 是组织 Pages 的树状层级结构。包含 `_index.{md,org,html}` 文件的目录即为 Section。

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

注意：`title` 留空时自动取目录名，根 Section 默认为 `index`。
`assets` 中的文件路径必须是相对当前 Section 目录的干净路径，不能使用绝对路径、`./` 或 `../`，并支持 glob 匹配，例如 `images/**/*.png`。附属资源会根据栏目最终的 `section.Path` 输出；如果 `section.Path` 是 `/blog.html`，则 `cover.png` 输出到 `/cover.png`，`media/cover.png` 输出到 `/media/cover.png`。

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

| 属性 | 说明 |
|------|------|
| `section.Title` | 栏目标题 |
| `section.Slug` | 栏目 slug |
| `section.Lang` | 语言代码 |
| `section.Path` | 相对 URL |
| `section.Permalink` | 绝对 URL |
| `section.Content` | `_index.md` 正文渲染内容 |
| `section.RawContent` | `_index.md` 原始内容 |
| `section.Pages` | 栏目下页面列表 |
| `section.Children` | 子栏目列表 |
| `section.Formats` | 其他输出格式 |
| `paginator` | 分页对象（启用分页时） |

## 分页变量

| 属性 | 说明 |
|------|------|
| `paginator.Path` | 当前分页链接 |
| `paginator.Permalink` | 当前分页绝对链接 |
| `paginator.PageNum` | 当前页码 |
| `paginator.Total` | 总页数 |
| `paginator.HasPrev()` | 是否有上一页 |
| `paginator.Prev.Path` | 上一页链接 |
| `paginator.Prev.Permalink` | 上一页绝对链接 |
| `paginator.HasNext()` | 是否有下一页 |
| `paginator.Next.Path` | 下一页链接 |
| `paginator.Next.Permalink` | 下一页绝对链接 |
| `paginator.All` | 所有分页对象 |
| `paginator.List` | 当前分页下的页面列表 |
