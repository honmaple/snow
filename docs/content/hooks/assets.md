---
title: "assets"
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
| `files` | 源文件 |
| `filters` | 过滤器（`cssmin`、`jsmin`） |
| `output` | 输出路径 |

`.scss`/`.sass` 文件自动编译为 CSS。

模板：

```html
{% assets css %}
<link rel="stylesheet" href="{{ config.base_url }}/{{ asset_url }}">
{% endassets %}
```
