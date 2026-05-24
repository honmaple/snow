---
title: "rewrite"
---

## Rewrite

重写 FrontMatter 字段。

```yaml
hooks:
  rewrite:
    enabled: true
    option:
      - src: "tag"
        dst: "tags"
        type: "list"
```

| 字段 | 说明 |
|------|------|
| `src` | 源字段名 |
| `dst` | 目标字段名 |
| `type` | `"list"` 按逗号分割为数组 |

