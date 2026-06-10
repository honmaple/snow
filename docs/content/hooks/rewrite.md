---
title: "rewrite"
weight: 50
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
| `src` | 源字段名，必填 |
| `dst` | 目标字段名，必填 |
| `type` | 可选；为空时原样复制，`"list"` 按逗号分割为数组 |

配置校验会在插件初始化时执行：缺少 `src`、缺少 `dst` 或使用未知 `type` 都会使构建失败。
