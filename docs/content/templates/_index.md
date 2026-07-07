---
title: "Templates"
weight: 50
sort_by: "weight"
---

Snow 使用基于 [Pongo2](https://github.com/flosch/pongo2) 的模板引擎，语法兼容 Django/Jinja2。

## 基本语法

```html
{# 注释 #}

<!-- 变量输出 -->
<h1>{{ page.Title }}</h1>
<p>{{ page.Date | date:"2006-01-02" }}</p>

<!-- 条件 -->
{% if page.Draft %}
  <span>草稿</span>
{% endif %}

<!-- 循环 -->
{% for page in pages %}
  <li><a href="{{ page.Path }}">{{ page.Title }}</a></li>
{% endfor %}

<!-- 翻译 -->
{% i18n "tags" %}
{{ _("共 %d 篇", 12) }}
```

## 全局变量

| 变量 | 说明 |
|------|------|
| `config` | 站点配置对象 |
| `pages` | 所有页面列表（Page） |
| `hidden_pages` | 隐藏页面列表 |
| `sections` | 所有栏目列表（Sections） |
| `taxonomies` | 所有分类列表（Taxonomies） |
| `page` | 当前页面（单页模板） |
| `section` | 当前栏目（栏目模板） |
| `term` | 当前分类项（Term 模板） |
| `taxonomy` | 当前分类（列表模板） |
| `terms` | `taxonomy.Terms` 引用（列表模板） |
| `paginator` | 分页对象 |

## 页面变量

`page` 对象属性：

| 属性 | 类型 | 说明 |
|------|------|------|
| `page.Title` | string | 标题 |
| `page.Slug` | string | URL slug |
| `page.Lang` | string | 语言 |
| `page.Date` | time.Time | 创建时间 |
| `page.Modified` | time.Time | 修改时间 |
| `page.Path` | string | 相对 URL |
| `page.Permalink` | string | 绝对 URL |
| `page.Summary` | string | 摘要 |
| `page.Content` | string | 渲染后 HTML |
| `page.RawContent` | string | 原始内容 |
| `page.WordCount` | int64 | 词数统计 |
| `page.ReadingTime` | int64 | 预计阅读时间（分钟） |
| `page.FrontMatter.{xxx}` | any | 自定义字段值 |
| `page.Formats` | Formats | 其他格式 |
| `page.Section` | *Section | 所属栏目 |
| `page.Ancestors()` | Sections | 从所属栏目到首页的栏目列表 |
| `page.Draft` | bool | 是否草稿 |
| `page.Hidden` | bool | 是否隐藏 |

## 栏目变量

| 属性 | 说明 |
|------|------|
| `section.Title` | 标题 |
| `section.Slug` | slug |
| `section.Lang` | 语言 |
| `section.Path` | 相对 URL |
| `section.Permalink` | 绝对 URL |
| `section.Content` | 正文 HTML |
| `section.RawContent` | `_index.md` 原始内容 |
| `section.WordCount` | 词数统计 |
| `section.ReadingTime` | 预计阅读时间（分钟） |
| `section.HiddenPages` | 隐藏页面列表 |
| `section.Pages` | 页面列表 |
| `section.Children` | 子栏目 |
| `section.Parent` | 父栏目 |
| `section.Ancestors()` | 从父栏目到首页的栏目列表 |
| `section.Formats` | 其他格式 |

## 分类变量

| 属性 | 说明 |
|------|------|
| `taxonomy.Name` | 分类名称 |
| `taxonomy.Lang` | 语言 |
| `taxonomy.Path` | 相对 URL |
| `taxonomy.Permalink` | 绝对 URL |
| `taxonomy.Terms` | Term 列表 |
| `term.Name` | Term 名称 |
| `term.Slug` | slug |
| `term.Path` | 相对 URL |
| `term.Permalink` | 绝对 URL |
| `term.Pages` | 页面列表 |
| `term.Children` | 子 Term |
| `term.Parent` | 父 Term |
| `term.Formats` | 其他格式 |
| `term.Taxonomy` | 所属 Taxonomy |

## 全局函数

| 函数 | 说明 |
|------|------|
| `get_pages([lang])` | 指定语言全部页面 |
| `get_hidden_pages([lang])` | 指定语言隐藏页面 |
| `get_sections([lang])` | 指定语言栏目列表 |
| `get_taxonomies([lang])` | 指定语言分类列表 |
| `get_page(path, [lang])` | 按路径获取页面 |
| `get_page_url(path, [lang])` | 页面绝对链接 |
| `get_section(path, [lang])` | 栏目 |
| `get_section_url(path, [lang])` | 栏目绝对链接 |
| `get_taxonomy(name, [lang])` | 分类 |
| `get_taxonomy_url(name, [lang])` | 分类绝对链接 |
| `get_taxonomy_term(taxonomy, name, [lang])` | 分类项 |
| `get_taxonomy_term_url(taxonomy, name, [lang])` | 分类项绝对链接 |
| `i18n(key)` | 翻译 |
| `T(key, args...)` | 格式化翻译 |
| `_(key, args...)` | 格式化翻译 |
| `dict(k, v, ...)` | 构造 map |
| `slice(v, ...)` | 构造列表 |
| `startsWith(s, prefix)` | 判断字符串前缀 |
| `load_data(path, format)` | 从 `data/` 或 URL 加载数据 |
| `newScratch()` | 创建临时模板存储 |

语言参数可选，省略则用当前语言。

## 页面列表方法

`pages` / `section.Pages` 等是 `Pages` 类型，支持链式操作：

### OrderBy

```html
{% for page in pages.OrderBy("date desc, title asc") %}
```

可用字段：`date`、`modified`、`title`、`weight` 及任意 FrontMatter 字段。

### GroupBy

```html
{% for group in pages.GroupBy("date:2006-01").OrderBy("name desc") %}
  <h2>{{ group.Name }}</h2>
  {% for page in group.Pages %}
    <li>{{ page.Title }}</li>
  {% endfor %}
{% endfor %}
```

`GroupBy` 返回 `PageGroups`，用于模板中的临时页面分组。参数可为 FrontMatter 字段名或 `date:{格式}`，`PageGroups.OrderBy` 支持 `name` 和 `count`。常用字段：`group.Name`、`group.Pages`、`group.Parent`、`group.Children`。

### Limit

```html
{% for page in pages.Limit(5) %}
```

## 过滤器

| 过滤器 | 说明 | 示例 |
|--------|------|------|
| `truncate:N` | 截取 N 字符 | `{{ text \| truncate:100 }}` |
| `date:"layout"` | 时间格式化 | `{{ page.Date \| date:"2006-01-02" }}` |
| `lower` | 转小写 | `{{ text \| lower }}` |
| `upper` | 转大写 | `{{ text \| upper }}` |
| `split:"sep"` | 分割字符串 | `{{ value \| split:"," }}` |
| `encrypt:"pw"` | 密码加密 | `{{ content \| encrypt:"123" }}` |
| `parser:"markdown"` | 使用已启用的内容解析器把字符串转为 HTML | `{{ text \| parser:"markdown" }}` |
| `unmarshal:"yaml"` | 解析 YAML/TOML/JSON 字符串 | `{{ text \| unmarshal:"yaml" }}` |
| `jsonify` | 转 JSON 字符串 | `{{ page.FrontMatter \| jsonify }}` |
| `absURL` | 转绝对 URL | `{{ "posts/" \| absURL }}` |
| `relURL` | 转相对 URL | `{{ "/posts/" \| relURL }}` |
| `slient` | 丢弃输出，仅保留函数副作用 | `{{ scratch.Set("k", 1) \| slient }}` |

Date 格式示例：`"2006-01-02"`、`"2006/01/02"`、`"January 2, 2006"`、`"Mon, 02 Jan 2006 15:04:05 -0700"` (RSS)。

## Shortcode

```html
<shortcode youtube id="xxx" />
<shortcode code lang="python">
print("hello")
</shortcode>
```

Shortcode 模板位于 `templates/shortcodes/`。模板中可用 `params`、`body`、`name`、`counter`、`current_lang`，页面内容提供 `page`，Section 内容提供 `section`。详细说明见 [短代码 (Shortcode)](/content/shortcodes/)。

## Assets

```html
{% assets css %}
<link rel="stylesheet" href="{{ config.base_url }}/{{ asset_url }}">
{% endassets %}
```

需启用 `hooks.assets`。

## Scratch 与数据

```html
{{ scratch.Set("count", 1) | slient }}
{{ scratch.Add("count", 2) | slient }}
{{ scratch.Get("count") }}

{% set data = load_data("links.yaml", "yaml") %}
```

`scratch` 是全局临时存储；需要局部存储时使用 `newScratch()`。`load_data` 可读取 `data/` 目录或 HTTP(S) URL，格式支持 `yaml`、`json`，其他格式返回原始字符串。

## 配置访问

```html
<title>{{ config.title }}</title>
<base href="{{ config.base_url }}">
<meta name="author" content="{{ config.author }}">
```
