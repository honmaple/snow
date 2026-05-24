# 输出格式

Snow 可为 Section、Page、Taxonomy Term 生成 RSS、Atom、JSON 等多种输出格式。

## 全局默认值

```yaml
formats:
  rss:
    template: "partials/rss.xml"
  atom:
    template: "partials/atom.xml"
```

定义全局默认模板，各内容类型可覆盖路径。

## Section 格式

通过 Section FrontMatter 配置：

```yaml
# _index.md
---
formats:
  rss:
    path: "posts/index.xml"
  atom:
    path: "posts/atom.xml"
  json:
    path: "posts/index.json"
    template: "custom.json"
---
```

- `path: ""` 禁用该格式
- 未设 `template` 时使用全局默认模板

## Page 格式

```yaml
---
formats:
  json:
    path: "api/articles/hello.json"
    template: "article.json"
---
```

## Taxonomy Term 格式

```yaml
taxonomies:
  tags:
    term:
      formats:
        atom:
          path: "tags/{term:slug}/atom.xml"
          template: "custom.atom.xml"
```

## 格式查找逻辑

1. FrontMatter 中的 `formats.{name}` map
2. `formats.{name}.path` — 输出路径（必须）
3. `formats.{name}.template` — 模板（可选，为空则从全局 `formats.{name}.template` 取值）
4. `path` 为空或 `template` 为空时跳过该格式

## 模板变量

| 内容类型 | 可用变量 |
|----------|----------|
| Section | `section`, `pages` |
| Page | `page` |
| Taxonomy Term | `term`, `pages`, `taxonomy` |

## RSS 模板示例

`templates/partials/rss.xml`：

```xml
<?xml version="1.0" encoding="utf-8"?>
<rss version="2.0">
  <channel>
    <title>{{ config.title }}</title>
    <link>{{ config.base_url }}</link>
    <description>{{ config.description }}</description>
    {% for page in pages %}
    <item>
      <title>{{ page.Title }}</title>
      <link>{{ page.Permalink }}</link>
      <pubDate>{{ page.Date | date:"Mon, 02 Jan 2006 15:04:05 -0700" }}</pubDate>
      <description>{{ page.Summary }}</description>
    </item>
    {% endfor %}
  </channel>
</rss>
```

## Atom 模板示例

```xml
<?xml version="1.0" encoding="utf-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
  <title>{{ config.title }}</title>
  <link href="{{ config.base_url }}/atom.xml" rel="self"/>
  <updated>{{ pages[0].Date | date:"2006-01-02T15:04:05Z07:00" }}</updated>
  <id>{{ config.base_url }}/</id>
  {% for page in pages %}
  <entry>
    <title>{{ page.Title }}</title>
    <link href="{{ page.Permalink }}"/>
    <id>{{ page.Permalink }}</id>
    <published>{{ page.Date | date:"2006-01-02T15:04:05Z07:00" }}</published>
    <content type="html">{{ page.Content }}</content>
  </entry>
  {% endfor %}
</feed>
```
