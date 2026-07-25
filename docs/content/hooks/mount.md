---
title: "mount"
weight: 5
---

## Mount

`mount` 可以把站点外部的本地文件或目录接入构建流程。例如把另一个项目的文档放到 `content/docs/project-name/` 下，或者把单个 CSS 文件放到 `static/style.css`。

```yaml
hooks:
  mount:
    enabled: true
    option:
      - source: "/tmp/project-name/docs"
        target: "content/docs/project-name"
        strategy: "mount"
      - source: "/tmp/project-name/static/style.css"
        target: "static/style.css"
```

| 字段 | 说明 |
|------|------|
| `source` | 本地文件或目录路径，必填 |
| `target` | 在站点中的目标路径，必填 |
| `strategy` | 合并策略，可选，默认 `mount` |

`target` 使用站点内路径，例如 `content/docs/demo`、`static/style.css` 或 `templates/partials/card.html`。它不能写成绝对路径，也不能包含 `.`、`./`、`..` 这类回退或自引用片段。

目录挂载后会把目录内容放到 `target` 下；文件挂载后会把该文件作为 `target` 使用。开发服务器不会监听外部 `source`，修改挂载源后需要重新构建或重启预览。

## 合并策略

| strategy | 说明 |
|----------|------|
| `mount` | 默认行为。挂载内容与原目录合并，同名文件使用挂载内容 |
| `base` | 挂载内容与原目录合并，同名文件使用原目录内容 |
| `override` | 目标路径只读取挂载内容，不再读取原目录内容 |

如果目标目录原本已经有文件，`mount` 和 `base` 会合并两边的内容；`override` 会让该目标完全以挂载内容为准。
