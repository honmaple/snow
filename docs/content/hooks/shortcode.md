---
title: "shortcode"
weight: 10
---

## Shortcode

在内容中使用可复用 HTML 片段。

```html
<shortcode youtube id="dQw4w9WgXcQ" />
```

```html
<shortcode code lang="python">
print("hello")
</shortcode>
```

Shortcode 模板放在 `templates/shortcodes/`，模板变量：`params.{key}`、`body`、`name`、`counter`、`current_lang`。页面内容提供 `page`，Section 内容提供 `section`。`params` 支持 `Get("key")`、`Pop("key")`、`String()`。
