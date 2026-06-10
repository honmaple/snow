---
title: "多语言"
weight: 60
---

Snow 内置多语言支持，通过三种方式区分内容语言。

## 配置

```yaml
# 默认语言
language: "zh"

languages:
  en:
    content_dir: "content/en"
    translations: "i18n/en.yaml"
  fr:
    translations: "i18n/fr.yaml"
```

`languages.{lang}` 可覆盖任何全局配置（`sections`、`pages`、`taxonomies`、`base_url` 等）。

## 语言检测

每个文件的语言按以下优先级确定：

1. FrontMatter 中的 `lang` 字段
2. 文件后缀（如 `hello.en.md` → `en`）
3. 站点默认语言

语言必须存在于 `languages` 配置中，否则回退到默认语言。

## 三种使用方式

### 方式一：按目录

```yaml
language: "zh"
content_dir: "content/zh"

languages:
  en:
    content_dir: "content/en"
```

### 方式二：按文件后缀

```
content/
├── _index.md              # 中文（默认语言）
├── _index.en.md           # 英文
├── about.md               # 中文
└── posts/
    ├── hello.md           # 中文
    └── hello.en.md        # 英文
```

### 方式三：FrontMatter

```yaml
---
title: "Hello World"
lang: "en"
---
```

## URL 映射

默认 URL 规则：

```
content/_index.md              → /index.html
content/_index.en.md           → /en/index.html
content/posts/hello.md         → /posts/hello/
content/posts/hello.en.md      → /en/posts/hello/
```

通过路径变量自定义：

```yaml
pages:
  posts:
    path: "articles/{date:%Y}/{date:%m}/{slug}/"

languages:
  en:
    pages:
      posts:
        path: "/english/articles/{date:%Y}/{slug}/"
```

## i18n

### 模板中使用

```html
{% i18n "tags" %}
{% T "共 %d 篇文章" 12 %}
{{ i18n("authors") }}
{{ T("hello %s", "world") }}
{{ _("共 %.2f", 3.14) }}
```

`i18n`、`T`、`_` 等价。

### 翻译文件

```yaml
# i18n/zh.yaml
---
- id: "authors"
  tr: "作者"
- id: "tags"
  tr: "标签"
```

```yaml
# i18n/en.yaml
---
- id: "authors"
  tr: "Authors"
- id: "tags"
  tr: "Tags"
```

### 配置翻译

```yaml
languages:
  en:
    translations: "i18n/en.yaml"
  zh:
    translations:
      - id: "authors"
        tr: "作者"
      - id: "tags"
        tr: "标签"
```

`translations` 可为文件路径或内联数组。主题 `i18n/` 目录下的文件也会自动加载。

## 跨语言函数

```html
{{ get_page("posts/hello", "en") }}
{{ get_section("posts", "en") }}
{{ get_taxonomy("tags", "en") }}
```

省略语言参数则使用当前语言。
