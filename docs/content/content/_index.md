---
title: "Content"
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

所有格式均支持 YAML 格式的 FrontMatter（页头元数据）。

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

## 内容导航

- [页面 (Page)](/content/pages) — 页面元数据、路径变量、模板变量
- [栏目 (Section)](/content/sections) — 栏目配置、分页、子栏目
- [分类系统 (Taxonomy)](/content/taxonomies) — 标签、分类、作者等
- [短代码 (Shortcode)](/content/shortcodes) — 在内容中插入可复用组件
- [多语言](/content/multilingual) — 国际化配置与使用
- [分页](/content/pagination) — Section 和 Taxonomy 级别的分页
- [输出格式](/content/formats) — RSS、Atom、JSON 等自定义格式
