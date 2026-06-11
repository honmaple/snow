---
title: "模板函数"
weight: 10
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
| `dict(k, v, ...)` | 构造 map |
| `slice(v, ...)` | 构造列表 |
| `startsWith(s, prefix)` | 判断字符串前缀 |
| `load_data(path, format)` | 从 `data/` 或 URL 加载数据 |
| `newScratch()` | 创建临时模板存储 |

## 页面列表方法

### OrderBy

```html
{% for page in pages.OrderBy("date desc, title asc") %}
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

### Reverse / FilterBy / OrderBy

```html
{% for page in pages.Reverse() %}
{% for page in pages.FilterBy("'go' in tags") %}
{% for page in pages.OrderBy("weight asc, date desc") %}
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
| `markdown` | Markdown 转 HTML | `{{ text \| markdown }}` |
| `org` | Org-mode 转 HTML | `{{ text \| org }}` |
| `parser:"yaml"` | 解析 YAML/TOML/JSON 字符串 | `{{ text \| parser:"yaml" }}` |
| `jsonify` | 转 JSON 字符串 | `{{ page.FrontMatter \| jsonify }}` |
| `absURL` | 转绝对 URL | `{{ "posts/" \| absURL }}` |
| `relURL` | 转相对 URL | `{{ "/posts/" \| relURL }}` |
| `slient` | 丢弃输出，仅保留函数副作用 | `{{ scratch.Set("k", 1) \| slient }}` |

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

`config` 在模板中是当前语言的配置快照。普通值可用点号访问，需要 viper 方法时可使用 `config.GetString(...)`。

## Scratch

```html
{{ scratch.Set("count", 1) | slient }}
{{ scratch.Add("count", 2) | slient }}
{{ scratch.Get("count") }}

{% set local = newScratch() %}
{{ local.Set("items", slice("a")) | slient }}
{{ local.Add("items", "b") | slient }}
{{ local.JSON("items") }}
```

支持方法：`Set`、`Get`、`GetOrSet`、`Add`、`JSON`。

## 数据加载

```html
{% set data = load_data("links.yaml", "yaml") %}
{% set remote = load_data("https://example.com/data.json", "json") %}
```

本地路径从站点或主题的 `data/` 目录读取；`format` 支持 `yaml`、`json`，其他格式按字符串返回。读取失败时返回 `nil` 并记录 warn 日志。

## Assets 块

```html
{% assets css %}
<link rel="stylesheet" href="{{ config.base_url }}/{{ asset_url }}">
{% endassets %}

{% assets files="scss/style.scss" sass_compiler="dartsass" filters="cssmin" output="css/style.min.css" %}
<link rel="stylesheet" href="{{ asset_url }}">
{% endassets %}
```

需在配置中启用 `hooks.assets`。`sass_compiler` 可使用 `libscss`、`dartsass`，只影响 `.scss` / `.sass` 文件；`filters` 可使用 `cssmin`、`jsmin`，用于合并后的处理。`dartsass` 需要本机可执行文件 `sass` 在 `PATH` 中可用，并支持 `sass --embedded`。

## Shortcode

```html
<shortcode youtube id="xxx" />
<shortcode code lang="python">
print("hello")
</shortcode>
```

模板位于 `templates/shortcodes/`。可用变量：`params.{key}`、`body`、`name`、`counter`。
