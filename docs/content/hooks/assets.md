---
title: "assets"
weight: 20
---
## Assets

静态资源处理：SCSS 编译、合并、压缩、图片缩放。

```yaml
hooks:
  assets:
    enabled: true
    option:
      css:
        files:
          - "scss/style.scss"
          - "css/custom.css"
        sass_compiler: "dartsass"
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
      cover:
        files:
          - "images/cover.jpg"
        filters:
          - name: "image"
            width: 1200
            height: 800
            fit: "inside"
            quality: 82
        output: "static/cover.jpg"
```

| 选项 | 说明 |
|------|------|
| `files` | 源文件，必填，支持 glob |
| `sass_compiler` | Sass 编译器，可选值：`libscss`、`dartsass`；只影响 `.scss` / `.sass` 文件 |
| `filters` | 过滤器，可使用字符串列表或对象步骤；可选值：`cssmin`、`jsmin`、`image` |
| `output` | 输出路径，必填 |

`.scss`/`.sass` 文件自动编译为 CSS。默认使用 `libscss`；配置 `sass_compiler: "dartsass"` 时改用 Dart Sass Embedded，需要本机可执行文件 `sass` 在 `PATH` 中可用，并支持 `sass --embedded`。

`cssmin`、`jsmin` 用于合并后的文本资源处理，不接受参数。`image` 用于单个图片文件，支持 `width`、`height`、`fit`、`quality` 参数；`fit` 可选 `inside`、`cover`、`fill`，默认 `inside`；`quality` 仅用于 JPEG，范围 `1..100`，默认 `85`。图片输出格式由 `output` 扩展名决定，支持 `.jpg`、`.jpeg`、`.png`、`.gif`；GIF 会按静态图处理，不保留动画。

使用 `image` filter 时，`files` 的 glob 最终必须只匹配一个图片文件，不会执行文本合并。模板内联 `assets` 标签只支持字符串形式的 `filters`，带参数的图片处理建议写在配置中。

配置校验会在插件初始化时执行：缺少 `files`、缺少 `output`、使用未知 `filters` 或非法 filter 参数都会使构建失败。

模板：

```html
{% assets css %}
<link rel="stylesheet" href="{{ config.base_url }}/{{ asset_url }}">
{% endassets %}
```
