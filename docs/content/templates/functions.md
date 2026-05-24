---
title: "模板函数"
---

## 内容获取函数

| 函数 | 说明 |
|------|------|
| `pages` | 当前语言全部页面 |
| `hidden_pages` | 当前语言隐藏页面 |
| `sections([lang])` | 栏目列表 |
| `taxonomies([lang])` | 分类列表 |
| `get_page(path, [lang])` | 获取页面对象 |
| `get_page_url(path, [lang])` | 获取页面绝对链接 |
| `get_section(path, [lang])` | 获取栏目对象 |
| `get_section_url(path, [lang])` | 获取栏目绝对链接 |
| `get_taxonomy(name, [lang])` | 获取分类对象 |
| `get_taxonomy_url(name, [lang])` | 获取分类绝对链接 |
| `get_taxonomy_term(taxonomy, name, [lang])` | 获取分类项 |
| `get_taxonomy_term_url(taxonomy, name, [lang])` | 获取分类项绝对链接 |

## 页面列表方法

### SortBy

```html
{% for page in pages.SortBy("date desc, title asc") %}
```

排序字段：`date`、`modified`、`title`、`weight` 及任意 FrontMatter 字段。

### GroupBy

```html
{% for group in pages.GroupBy("date:2006-01").OrderBy("name desc") %}
  <h2>{{ group.Name }}</h2>
  {% for page in group.List %}
    <li>{{ page.Title }}</li>
  {% endfor %}
{% endfor %}
```

分组支持 FrontMatter 字段或 `date:{Go格式}`。

### Limit

```html
{% for page in pages.Limit(10) %}
```

### First / Last

```html
{{ pages.First() }}
{{ pages.Last() }}
```

### Related

获取相关（前/后）文章：

```html
{{ pages.Related(page).Prev }}
{{ pages.Related(page).Next }}
{{ pages.Related(page).HasPrev() }}
{{ pages.Related(page).HasNext() }}
```

即 `page.Prev`、`page.Next` 的来源。

## i18n 翻译函数

| 函数 | 说明 |
|------|------|
| `i18n(key)` | 翻译字符串 |
| `T(key, args...)` | 格式化翻译 (`%d`, `%s`, `%f`) |
| `_(key, args...)` | 同上 |

## 过滤器

| 过滤器 | 说明 | 示例 |
|--------|------|------|
| `truncate:N` | 截取 N 字符 | `{{ text \| truncate:100 }}` |
| `date:"layout"` | Go 时间格式 | `{{ page.Date \| date:"2006-01-02" }}` |
| `lower` | 转小写 | `{{ text \| lower }}` |
| `upper` | 转大写 | `{{ text \| upper }}` |
| `split:sep` | 分割 | `{{ "a,b" \| split:"," }}` |
| `encrypt:"pw"` | 加密 (需 hooks.encrypt) | `{{ page.Content \| encrypt:"123" }}` |

Date 格式参考：

| 格式 | 输出示例 | 用途 |
|------|----------|------|
| `2006-01-02` | `2024-01-15` | 日期 |
| `15:04:05` | `20:35:00` | 时间 |
| `January 2, 2006` | `January 15, 2024` | 完整日期 |
| `Mon, 02 Jan 2006 15:04:05 -0700` | RSS pubDate |
| `2006-01-02T15:04:05Z07:00` | ISO 8601 |

## 配置访问

```html
{{ config.title }}
{{ config.base_url }}
{{ config.author }}
{{ config.description }}
{{ config.GetString("params.my_key") }}
```

## Assets 块

```html
{% assets css %}
<link rel="stylesheet" href="{{ config.base_url }}/{{ asset_url }}">
{% endassets %}

{% assets files="scss/style.scss" filters="cssmin" output="css/style.min.css" %}
<link rel="stylesheet" href="{{ asset_url }}">
{% endassets %}
```

需在配置中启用 `hooks.assets`。

## Shortcode

```html
<shortcode youtube id="xxx" />
<shortcode code lang="python">
print("hello")
</shortcode>
```

模板位于 `templates/shortcodes/`。可用变量：`params.{key}`、`body`、`name`、`counter`。
