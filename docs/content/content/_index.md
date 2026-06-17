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

## 内容解析

Snow 内置 Markdown、Org-mode、HTML 解析器，并提供可选的 `niklasfasching` Org-mode 解析器。解析器启用方式、元数据来源、摘要分隔符和 `Toc` 行为详见 [解析器 (Parser)](/content/parsers)。

## FrontMatter

FrontMatter 是内容文件开头或 HTML `<head>` 中的页面元数据，定义页面的标题、日期、标签等。Markdown 常用 YAML 元数据块：

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

## 内容导航

- [解析器 (Parser)](/content/parsers) — Markdown、Org-mode、niklasfasching、HTML 的解析规则
- [页面 (Page)](/content/pages) — 页面元数据、路径变量、模板变量
- [栏目 (Section)](/content/sections) — 栏目配置、分页、子栏目
- [附件资源 (Assets)](/content/assets) — Page Bundle 和 Section 附件复制规则
- [分类系统 (Taxonomy)](/content/taxonomies) — 标签、分类、作者等
- [短代码 (Shortcode)](/content/shortcodes) — 在内容中插入可复用组件
- [多语言](/content/multilingual) — 国际化配置与使用
- [分页](/content/pagination) — Section 和 Taxonomy 级别的分页
- [输出格式](/content/formats) — RSS、Atom、JSON 等自定义格式
