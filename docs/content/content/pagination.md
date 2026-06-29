---
title: "分页 (Pagination)"
weight: 40
---

Snow 支持 Section 和 Taxonomy Term 级别的分页。

## Section 分页

通过 FrontMatter 或配置启用：

```yaml
sections:
  posts:
    paginate: 10
    paginate_filter_by: ""
```

| 配置项 | 说明 |
|--------|------|
| `paginate` | 每页条目数，`0` 不分页（默认 `0`） |
| `paginate_path` | 分页路径模板，不设置时自动选择 |
| `paginate_filter_by` | 分页前过滤 |

当 `paginate_path` 未设置或为空字符串时，分页器会根据 Section/Term 的 `path` 选择默认值：目录路径（pretty）使用 `page/{number}/`，文件路径（ugly）使用 `{name}{number}{extension}`。

## Taxonomy Term 分页

```yaml
taxonomies:
  tags:
    term:
      paginate: 10
      paginate_filter_by: ""
```

## 分页路径变量

| 变量 | 说明 | 示例 |
|------|------|------|
| `{name}` | 源路径的文件名 | `index` |
| `{extension}` | 源路径的扩展名 | `.html` |
| `{number}` | 页码，第一页为 `1` | `1`, `2`, `3` |

分页路径基于 Section/Term 的 `path` 值拼接。

第一页始终使用 Section/Term 自身的 `path`。`paginate_path` 只用于第二页及之后的分页路径。

如果 Section/Term 的 `path` 以 `/` 结尾，会按目录路径处理后续分页：`{name}` 为 `index`，`{extension}` 为 `.html`。例如 `/posts/` 搭配 `{name}{number}{extension}` 会生成 `/posts/`、`/posts/index2.html`。

如果 `path` 是文件路径，后续分页会使用该文件名和扩展名。比如 `/posts.html` 搭配 `{name}-{number}{extension}` 会生成 `/posts.html`、`/posts-2.html`。

## 示例

```yaml
# 示例一：沿用源路径结构
path: "posts/"
paginate_path: "{name}{number}{extension}"

# → 第一页: /posts/
# → 第二页: /posts/index2.html

# 示例二：使用子目录
path: "posts/"
paginate_path: "page/{number}{extension}"

# → 第一页: /posts/
# → 第二页: /posts/page/2.html

# 示例三：留空 paginate_path，pretty 路径使用目录 fallback
path: "posts/"
paginate_path: ""

# → 第一页: /posts/
# → 第二页: /posts/page/2/

# 示例四：留空 paginate_path，ugly 路径沿用文件名
path: "posts.html"
paginate_path: ""

# → 第一页: /posts.html
# → 第二页: /posts2.html
```

## 模板变量

| 属性 | 说明 |
|------|------|
| `paginator.Path` | 当前分页链接 |
| `paginator.Permalink` | 当前分页绝对链接 |
| `paginator.PageNum` | 当前页码 |
| `paginator.Total` | 总页数 |
| `paginator.HasPrev()` | 有上一页 |
| `paginator.Prev.Path` | 上一页链接 |
| `paginator.Prev.Permalink` | 上一页绝对链接 |
| `paginator.HasNext()` | 有下一页 |
| `paginator.Next.Path` | 下一页链接 |
| `paginator.Next.Permalink` | 下一页绝对链接 |
| `paginator.First()` | 第一页分页对象 |
| `paginator.Last()` | 最后一页分页对象 |
| `paginator.Page(n)` | 指定页码的分页对象 |
| `paginator.Pagers` | 所有分页 |
| `paginator.Pages` | 当前分页页面列表 |

例：

```html
{% if paginator.HasPrev() %}
<a href="{{ paginator.Prev.Permalink }}">上一页</a>
{% endif %}

<span>{{ paginator.PageNum }} / {{ paginator.Total }}</span>

{% if paginator.HasNext() %}
<a href="{{ paginator.Next.Permalink }}">下一页</a>
{% endif %}
```
