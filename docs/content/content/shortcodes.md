---
title: "短代码 (Shortcode)"
weight: 50
---

Shortcode 用于在内容中插入可复用模板。当前实现由 `hooks.shortcode` 提供，作用于 Page 和 Section 的 `Content`、`Summary`。

## 启用

```yaml
hooks:
  shortcode:
    enabled: true
```

## 模板位置

Shortcode 模板从 `templates/shortcodes/` 加载，支持下面几种布局：

- `templates/shortcodes/{name}.html`
- `templates/shortcodes/{name}.tpl`
- `templates/shortcodes/{name}/index.html`
- `templates/shortcodes/{name}/index.tpl`

例如创建 `templates/shortcodes/bilibili.html`：

```html
<div class="shortcode-bilibili">
  <iframe
    src="https://player.bilibili.com/player.html?bvid={{ params.id }}&page=1"
    scrolling="no"
    border="0"
    frameborder="no"
    allowfullscreen="true"
  >
  </iframe>
</div>
```

## 内容中使用

推荐使用 HTML 形式的 shortcode：

```html
<shortcode bilibili id="BV1yB4cz8E9y" />
```

带 body 的形式：

```html
<shortcode notice type="warning">
  **注意：** 此功能正在开发中。
</shortcode>
```

Markdown 开启 `markups.markdown.directive_blocks: true` 后，也可以使用 `:::shortcode`：

````markdown
:::shortcode notice type=info
这里的 **Markdown** 会继续解析，并作为 shortcode body 传入。

- static-site
- go
:::
````

`:::shortcode` 会先把块内 Markdown 解析成 HTML，再作为 `body` 传入 shortcode。需要在 body 中保留原始 YAML、TOML 或其它未解析文本时，使用普通 HTML 形式的 `<shortcode ...>...</shortcode>`。

## 模板变量

| 变量 | 说明 |
|------|------|
| `params.{key}` | 传入的参数 |
| `params.Get("key")` | 读取参数值 |
| `params.Pop("key")` | 读取并移除参数，适合取走特定参数后透传剩余属性 |
| `params.String()` | 将剩余参数渲染为稳定排序的 HTML 属性字符串 |
| `body` | 闭合 shortcode 的内部内容；若包含嵌套 shortcode，会先渲染最内层 shortcode，再把渲染后的 HTML 作为外层 shortcode 的 body 传入 |
| `name` | 当前 shortcode 名称 |
| `counter` | 当前内容中此名称 shortcode 的出现次数，从 `0` 开始递增 |
| `current_lang` | 当前内容语言 |
| `page` | 当前 Page；只在页面内容中提供 |
| `section` | 当前 Section；只在 Section 内容中提供 |

`params.String()` 会按 key 排序输出属性；值为空字符串时输出布尔属性。例如：

```html
{% set type = params.Pop("type") %}
<div class="notice notice-{{ type }}" {{ params.String() }}>
  {{ body }}
</div>
```

## 行为说明

- Shortcode 作用于 Page 和 Section 的 `Content`、`Summary`。
- 嵌套 shortcode 按从内到外的顺序渲染；外层 shortcode 接收到的是内层渲染后的 HTML。
- 执行模板失败时会记录 warning，并保留原始 shortcode 标签与 body。
- 未闭合的 shortcode 会记录 warning，并保留已解析到的原始内容。

## 示例

### YouTube 嵌入

`templates/shortcodes/youtube.html`：

```html
<div class="shortcode-youtube">
  <iframe
    src="https://www.youtube.com/embed/{{ params.id }}"
    frameborder="0"
    allowfullscreen
  >
  </iframe>
</div>
```

使用：

```html
<shortcode youtube id="dQw4w9WgXcQ" />
```

### 代码块

`templates/shortcodes/code.html`：

```html
{% if params.lang %}
<pre><code class="language-{{ params.lang }}">{{ body }}</code></pre>
{% else %}
<pre><code>{{ body }}</code></pre>
{% endif %}
```

使用：

```html
<shortcode code lang="python">
def hello():
    print("Hello, Snow!")
</shortcode>
```
