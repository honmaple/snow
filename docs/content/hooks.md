# 插件 (Hooks)

Snow 的插件系统通过 Hooks 机制在内容处理流程中插入自定义逻辑。

## 查看插件

```bash
snow hooks
# 输出: assets(enabled), encrypt(enabled), filter, minify, pelican(enabled), rewrite(enabled), shortcode(enabled)
```

## 配置

```yaml
hooks:
  assets:
    enabled: true
  encrypt:
    enabled: true
    weight: 2
    option:
      password: "123456"
  shortcode:
    enabled: true
    weight: 1
  minify:
    enabled: false
    option:
      html: true
      css: true
      js: true
  rewrite:
    enabled: false
    option:
      - src: "tag"
        dst: "tags"
        type: "list"
  filter:
    option:
      page_filter: ""
```

`weight` 控制执行顺序，越大越优先。默认启用：`assets`、`encrypt`、`shortcode`。

## 内置插件

| 插件 | 默认启用 | 说明 |
|------|----------|------|
| shortcode | ✅ | 内容中嵌入可复用组件 |
| encrypt | ✅ | 内容加密 |
| assets | ✅ | 静态资源处理 |
| pelican | ❌ | 文档格式转换 |
| rewrite | ❌ | FrontMatter 重写 |
| minify | ❌ | 输出压缩 |
| filter | ❌ | 页面筛选 |

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

Shortcode 模板放在 `templates/shortcodes/`，模板变量：`params.{key}`、`body`、`name`、`counter`。

---

## Encrypt

内容加密。

```yaml
hooks:
  encrypt:
    enabled: true
    weight: 2
    option:
      password: "默认密码"
```

### 全篇加密

```yaml
---
password: "123456"
---
```

### 局部加密

```html
<shortcode encrypt password="123456">
加密内容
</shortcode>
```

### 模板中

```html
{{ page.Content | encrypt:"123456" }}
```

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

---

## Minify

HTML/CSS/JS 输出压缩。

```yaml
hooks:
  minify:
    enabled: true
    option:
      html: true
      css: true
      js: true
```

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

---

## Filter

内容筛选。

```yaml
hooks:
  filter:
    option:
      page_filter: "'emacs' in tags and not draft"
```

---

## Pelican

文档格式转换。

```yaml
hooks:
  pelican:
    enabled: true
```
