---
title: "目录结构"
weight: 20
---

Snow 站点的标准文件布局：

```
mysite/
├── config.yaml          # 站点配置
├── content/             # 内容目录（固定名）
│   ├── _index.md        # 根 Section
│   ├── about.md         # Page
│   └── posts/           # Section
│       ├── _index.md    # Section 内容
│       ├── hello.md     # Page
│       └── tutorials/   # 子 Section
│           └── _index.md
├── static/              # 静态文件（固定名，原样复制到输出）
├── templates/           # 站点自定义模板（覆盖主题模板）
├── themes/              # 主题目录（固定名）
│   └── snow/            # 主题名称 = theme
│       ├── theme.yaml
│       ├── templates/
│       ├── static/
│       └── i18n/
├── assets/              # 需启用 hooks.assets
│   ├── js/
│   ├── css/
│   └── scss/
└── i18n/                # 翻译文件
    ├── en.yaml
    └── zh.yaml
```

## 固定目录

| 目录 | 说明 |
|------|------|
| `content/` | 内容目录 |
| `static/` | 静态文件目录，原样复制到输出 |
| `templates/` | 站点模板目录，优先于主题模板 |
| `themes/` | 主题目录，`theme` 对应 `themes/{name}/` |
| `assets/` | assets hook 默认读取的资源目录 |
| `data/` | data 模板扩展默认读取的数据目录 |
| `i18n/` | i18n 模板扩展默认读取的翻译目录 |
| `output/` | 默认构建输出目录，可通过 `output_dir` 修改 |

核心目录 `content`、`static`、`templates`、`themes` 使用固定名称；扩展目录（如 `assets`、`data`、`i18n`）由各扩展按需交给虚拟文件系统读取。

`mount` hook 可把外部文件或目录挂载到这些虚拟路径中。默认策略为 `mount`，同名文件使用挂载内容；也可以配置 `base` 让原目录优先，或配置 `override` 让目标路径完全由挂载内容覆盖。

## 内容目录详解

### 识别规则

| 文件/目录 | 识别为 |
|-----------|--------|
| `*.md` / `*.org` / `*.html` | Page |
| `_index.{md,org,html}` | Section 标记 |
| `index.{md,org,html}` | Page Bundle 标记 |
| `_*` / `.*` 开头 | 默认忽略 |

`.html` 内容文件需要启用 `markups.html.enabled: true` 后才会被解析；默认启用的是 Markdown 和 Org-mode。

### Page

普通文件（`.md`、`.org`、`.html`）即为 Page。输出路径由 `pages.{type}.path` 决定。

### Section

包含 `_index.*` 的目录为 Section。Section 可嵌套形成层级结构。

### Page Bundle

包含 `index.{md,org,html}` 的目录视为一个页面整体，目录内其他文件为附属资源。

### 忽略规则

- 以 `_` 或 `.` 开头的文件/目录（`_index.*` 除外）
- `ignored_content` 匹配的内容

```yaml
ignored_content:
  - "drafts/"
  - "private/*"
```
