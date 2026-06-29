---
title: "分类系统 (Taxonomy)"
weight: 30
---

Snow 根据 Page FrontMatter 字段自动生成分类页面。类似 Hugo 的 Taxonomy 或 Zola 的 Taxonomies。

## 工作原理

只需在 `taxonomies` 下列出字段名，Snow 会扫描所有 Page，按该字段值分组，自动生成分类列表页和 Term 详情页。

Term 支持层级结构：`categories` 中值为 `Programming/Go` 时，自动生成父 term `Programming` 和子 term `Go`。

## 配置

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

只需列出名称即可启用。

### 配置项

**分类级别**（`taxonomies.{name}.{key}`）：

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| `path` | `{taxonomy}/` | 分类列表页路径 |
| `path_style` | `none` | 路径后处理，见 [Path Style 配置](/configuration/#path-style-配置) |
| `template` | — | 自定义模板 |
| `sort_by` | `name` | Term 排序（`"name"` 或 `"count"`） |

**Term 级别**（`taxonomies.{name}.term.{key}`）：

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| `term.path` | `{taxonomy}/{term:slug}/` | Term 页面路径 |
| `term.path_style` | `none` | 路径后处理，见 [Path Style 配置](/configuration/#path-style-配置) |
| `term.template` | — | Term 页面模板 |
| `term.sort_by` | `date desc` | Term 下页面排序 |
| `term.filter_by` | — | Term 下页面筛选 |
| `term.paginate` | — | 分页数，不设不分页 |
| `term.paginate_path` | 自动 | 分页路径模板 |
| `term.paginate_filter_by` | — | 分页前筛选 |

配置查找：`taxonomies.{name}.{key}` → `taxonomies._default.{key}`

`term.paginate_path` 未设置或为空字符串时会按 Term 输出路径类型选择默认值；分页路径变量和模板对象见 [分页](/content/pagination)。

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
| `term.Pages` | Term 下页面列表 |
| `term.Children` | 子 Term |
| `term.Parent` | 父 Term |
| `term.Formats` | 其他输出格式 |
| `term.Taxonomy` | 所属 Taxonomy 对象 |
| `paginator` | 分页对象 |

Term 模板中不再注入 `pages` 作为 `term.Pages` 的别名；需要当前 Term 下页面列表时直接使用 `term.Pages`。

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
  {% for page in group.Pages %}
    <li>{{ page.Title }}</li>
  {% endfor %}
{% endfor %}
```
