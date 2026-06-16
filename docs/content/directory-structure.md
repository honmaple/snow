---
title: "目录结构"
weight: 20
---

Snow 站点的标准文件布局：

```
mysite/
├── config.yaml          # 站点配置
├── content/             # 内容目录（content_dir）
│   ├── _index.md        # 根 Section
│   ├── about.md         # Page
│   └── posts/           # Section
│       ├── _index.md    # Section 内容
│       ├── hello.md     # Page
│       └── tutorials/   # 子 Section
│           └── _index.md
├── static/              # 静态文件（static_dir）
├── templates/           # 站点自定义模板（覆盖主题模板）
├── themes/              # 主题目录（theme_dir）
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

## 可配置目录

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| `content_dir` | `content` | 内容目录 |
| `static_dir` | `static` | 静态文件目录，原样复制到输出 |
| `output_dir` | `output` | 构建输出 |
| `theme_dir` | `themes` | 主题目录 |
| `theme` | — | 主题名称，对应 `themes/` 下子目录 |

## 内容目录详解

### 识别规则

| 文件/目录 | 识别为 |
|-----------|--------|
| `*.md` / `*.org` / `*.html` | Page |
| `_index.{md,org,html}` | Section 标记 |
| `index.{md,org,html}` | Page Bundle 标记 |
| `_*` / `.*` 开头 | 默认忽略 |

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

