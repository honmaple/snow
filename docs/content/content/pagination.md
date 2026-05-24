---
title: "分页"
---

Snow 支持 Section 和 Taxonomy Term 级别的分页。

## Section 分页

通过 FrontMatter 或配置启用：

```yaml
sections:
  posts:
    paginate: 10
    paginate_path: "{name}{number:optional}{extension}"
    paginate_filter_by: ""
```

| 配置项 | 说明 |
|--------|------|
| `paginate` | 每页条目数，`0` 不分页 （默认 `10`） |
| `paginate_path` | 分页路径模板（默认 `{name}{number:optional}{extension}`） |
| `paginate_filter_by` | 分页前过滤 |

## Taxonomy Term 分页

```yaml
taxonomies:
  tags:
    term:
      paginate: 10
      paginate_path: "{name}{number:optional}{extension}"
      paginate_filter_by: ""
```

## 分页路径变量

| 变量 | 说明 | 示例 |
|------|------|------|
| `{name}` | 源路径的文件名 | `index` |
| `{extension}` | 源路径的扩展名 | `.html` |
| `{number}` | 页码，第一页为 `1` | `1`, `2`, `3` |
| `{number:optional}` | 页码，第一页为空 | `2`, `3` (第一页空) |

分页路径基于 Section/Term 的 `path` 值拼接。

## 示例

```yaml
# 示例一：沿用源路径结构
path: "posts/"
paginate_path: "{name}{number:optional}{extension}"

# → 第一页: /posts/index.html
# → 第二页: /posts/index2.html

# 示例二：使用子目录
path: "posts/"
paginate_path: "page/{number}{extension}"

# → 第一页: /posts/page/1.html
# → 第二页: /posts/page/2.html
```

## 模板变量

| 属性 | 说明 |
|------|------|
| `paginator.Path` | 当前分页链接 |
| `paginator.PageNum` | 当前页码 |
| `paginator.Total` | 总页数 |
| `paginator.HasPrev()` | 有上一页 |
| `paginator.Prev.URL` | 上一页链接 |
| `paginator.HasNext()` | 有下一页 |
| `paginator.Next.URL` | 下一页链接 |
| `paginator.All` | 所有分页 |
| `paginator.List` | 当前分页页面列表 |

例：

```html
{% if paginator.HasPrev() %}
<a href="{{ paginator.Prev.URL }}">上一页</a>
{% endif %}

<span>{{ paginator.PageNum }} / {{ paginator.Total }}</span>

{% if paginator.HasNext() %}
<a href="{{ paginator.Next.URL }}">下一页</a>
{% endif %}
```
