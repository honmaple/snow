---
title: "filter"
weight: 40
---

## Filter

内容筛选。

```yaml
hooks:
  filter:
    enabled: true
    option:
      page_filter: "'emacs' in tags and not draft"
```

`page_filter` 使用 Pongo2 表达式语法。插件初始化时会预编译该表达式；如果语法错误，构建会直接失败，而不是等到页面处理阶段再忽略。
