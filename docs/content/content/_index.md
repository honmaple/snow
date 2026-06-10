---
title: "Content"
weight: 40
sort_by: "weight"
---

Snow 的内容系统围绕三个核心概念构建：

| 概念 | 说明 | 识别方式 |
|------|------|----------|
| [**Page**](/content/pages) | 单篇文章/页面 | `.md`、`.org`、`.html` 文件 |
| [**Section**](/content/sections) | 栏目/目录，组织 Pages | 包含 `_index.*` 的目录 |
| [**Taxonomy**](/content/taxonomies) | 标签/分类/作者等分类系统 | FrontMatter 字段自动生成 |

## 支持的格式

Snow 支持三种内容格式，每种使用不同的解析引擎：

| 格式 | 扩展名 | 解析引擎 |
|------|--------|----------|
| Markdown | `.md` | goldmark |
| Org-mode | `.org` | org-golang |
| HTML | `.html` | 内置解析器 |

Markdown 支持 YAML (`---`) 与 TOML (`+++`) FrontMatter；Org-mode 支持 `:PROPERTIES:` drawer 和文件开头的 `#+KEY:` 元数据；HTML 会从 `<title>`、`<meta name="..." content="...">`、`<link href="...">`、`<script src="...">` 提取元数据。

## FrontMatter

FrontMatter 是文件开头的 YAML 元数据块，定义页面的标题、日期、标签等：

```markdown
---
title: "我的文章"
date: 2024-01-15
tags:
  - go
  - web
categories:
  - Programming/Go
draft: false
---
```

更多 FrontMatter 字段详见 [页面 (Page)](/content/pages)。

## 解析限制

Markdown 与 Org-mode 解析器支持最长约 1MB 的单行内容。Markdown FrontMatter 必须使用同类型 fence 闭合；Org-mode 的 `:PROPERTIES:` drawer 必须以 `:END:` 闭合，否则构建会报错。

## 内容导航

- [页面 (Page)](/content/pages) — 页面元数据、路径变量、模板变量
- [栏目 (Section)](/content/sections) — 栏目配置、分页、子栏目
- [分类系统 (Taxonomy)](/content/taxonomies) — 标签、分类、作者等
- [短代码 (Shortcode)](/content/shortcodes) — 在内容中插入可复用组件
- [多语言](/content/multilingual) — 国际化配置与使用
- [分页](/content/pagination) — Section 和 Taxonomy 级别的分页
- [输出格式](/content/formats) — RSS、Atom、JSON 等自定义格式
