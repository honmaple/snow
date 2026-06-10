---
title: "assets"
weight: 20
---
## Assets

静态资源处理：SCSS 编译、合并、压缩。

```yaml
hooks:
  assets:
    enabled: true
    option:
      css:
        files:
          - "scss/style.scss"
          - "css/custom.css"
        filters:
          - "cssmin"
        output: "static/style.min.css"
      js:
        files:
          - "js/main.js"
          - "js/theme.js"
        filters:
          - "jsmin"
        output: "static/script.min.js"
```

| 选项 | 说明 |
|------|------|
| `files` | 源文件，必填，支持 glob |
| `filters` | 过滤器，可选值：`cssmin`、`jsmin`、`libscss` |
| `output` | 输出路径，必填 |

`.scss`/`.sass` 文件自动编译为 CSS。

配置校验会在插件初始化时执行：缺少 `files`、缺少 `output` 或使用未知 `filters` 都会使构建失败。

模板：

```html
{% assets css %}
<link rel="stylesheet" href="{{ config.base_url }}/{{ asset_url }}">
{% endassets %}
```
