---
title: "分类系统 (Taxonomy)"
weight: 30
---

Snow 根据 Page FrontMatter 字段自动生成分类页面。类似 Hugo 的 Taxonomy 或 Zola 的 Taxonomies。

## 工作原理

只需在 `taxonomies` 下列出字段名，Snow 会扫描所有 Page，按该字段值分组，自动生成分类列表页和 Term 详情页。

Term 支持层级结构：=categories= 中值为 =Programming/Go= 时，自动生成父 term =Programming= 和子 term =Go=。

## 配置

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

只需列出名称即可启用。

### 配置项

**分类级别**（`taxonomies.{name}.{key}`）：

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| `path` | `{taxonomy}/` | 分类列表页路径 |
| `template` | — | 自定义模板 |
| `sort_by` | `name` | Term 排序（`"name"` 或 `"count"`） |

**Term 级别**（`taxonomies.{name}.term.{key}`）：

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| `term.path` | `{taxonomy}/{term:slug}/` | Term 页面路径 |
| `term.template` | — | Term 页面模板 |
| `term.sort_by` | `date desc` | Term 下页面排序 |
| `term.filter_by` | — | Term 下页面筛选 |
| `term.paginate` | — | 分页数，不设不分页 |
| `term.paginate_path` | `{name}{number:optional}{extension}` | 分页路径模板 |
| `term.paginate_filter_by` | — | 分页前筛选 |

配置查找：`taxonomies.{name}.{key}` → `taxonomies._default.{key}`

## FrontMatter 中使用

```yaml
---
tags:
  - go
  - web
  - tutorial
categories:
  - Programming/Go
authors: honmaple
---
```

`/` 分隔的值自动形成层级结构。

## 路径变量

### 分类列表页

| 变量 | 说明 |
|------|------|
| `{lang}` | 语言代码 |
| `{lang:optional}` | 语言代码，默认语言为空 |
| `{taxonomy}` | 分类名称 |

### Term 页面

| 变量 | 说明 |
|------|------|
| `{lang}` | 语言代码 |
| `{lang:optional}` | 语言代码，默认语言空 |
| `{taxonomy}` | 分类名称 |
| `{term}` | Term 原始名称 |
| `{term:slug}` | Term slug 化 |

## 模板查找

### 分类列表页

1. `taxonomies.{name}.template` 配置值
2. `{name}/list.html`
3. `taxonomy_list.html`

### Term 页面

1. `taxonomies.{name}.term.template` 配置值
2. `{name}/single.html`
3. `taxonomy_single.html`

## 模板变量

### 分类列表模板

| 属性 | 说明 |
|------|------|
| `taxonomy.Name` | 分类名称 |
| `taxonomy.Lang` | 语言 |
| `taxonomy.Path` | 相对 URL |
| `taxonomy.Permalink` | 绝对 URL |
| `taxonomy.Terms` | 所有 Term 列表 |
| `terms` | `taxonomy.Terms` 别名 |

### Term 模板

| 属性 | 说明 |
|------|------|
| `term.Name` | Term 名称 |
| `term.Slug` | Term slug |
| `term.Path` | 相对 URL |
| `term.Permalink` | 绝对 URL |
| `term.List` | Term 下页面列表 |
| `term.Children` | 子 Term |
| `term.Parent` | 父 Term |
| `term.Formats` | 其他输出格式 |
| `term.Taxonomy` | 所属 Taxonomy 对象 |
| `pages` | `term.Pages` 引用 |
| `paginator` | 分页对象 |

## 时间归档

按年月归档：

```yaml
taxonomies:
  "date:2006/01":
    sort_by: "name desc"
    path: "archives/"
    template: "archives.html"
    term:
      path: "archives/{term}/"
      template: "period_archives.html"
```

格式 `date:2006/01` 指按年月分组。生成链接如 `/archives/2024/01/`。

或在模板中手动分组：

```html
{% for group in pages.GroupBy("date:2006-01").OrderBy("name desc") %}
  <h2>{{ group.Name }}</h2>
  {% for page in group.List %}
    <li>{{ page.Title }}</li>
  {% endfor %}
{% endfor %}
```
