---
title: "alias"
weight: 80
---

## Alias

`alias` 用于为旧 URL 生成重定向页面，适合迁移旧站点路径。

```yaml
hooks:
  alias:
    enabled: true
    option:
      - "/old-url/:/new-url/"
      - "/old-post.html:/posts/new-post/"
```

每条配置使用 `旧路径:新路径`。旧路径会写入输出目录，因此必须是站点内路径；不能包含查询参数、锚点或反斜杠。新路径可以是站点内路径，也可以是完整外部 URL。

如果旧路径以 `/` 结尾，Snow 会生成对应的 `index.html`；如果没有文件扩展名，也会按目录路径处理。
