# 短代码 (Shortcode)

Shortcode 用于在内容中快速插入可复用的 HTML 组件，如视频嵌入、代码片段等。需要在配置中启用 `hooks.shortcode`。

## 启用

```yaml
hooks:
  shortcode:
    enabled: true
    weight: 1
```

## 创建 Shortcode

在 `templates/shortcodes/` 目录下创建模板文件。以 Bilibili 视频嵌入为例，创建 `templates/shortcodes/bilibili.html`：

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

### 模板内可用变量

| 变量 | 说明 |
|------|------|
| `params.{key}` | 传入的参数 |
| `body` | Shortcode 闭合标签内的内容 |
| `name` | Shortcode 的名称 |
| `counter` | 当前文章中此 Shortcode 的使用次数 |

## 使用 Shortcode

在 Markdown/Org-mode 内容中使用，有两种语法：

```html
<!-- 自闭合 -->
<bilibili id="BV1yB4cz8E9y" />

<!-- 标准模板格式 -->
<shortcode bilibili id="BV1yB4cz8E9y" />
```

带 body 内容的用法：

```html
<shortcode notice type="warning">
  **注意：** 此功能正在开发中。
</shortcode>
```

在模板中访问：

```html
<div class="notice notice-{{ params.type }}">
  {{ body }}
</div>
```

## 示例：YouTube 嵌入

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
<youtube id="dQw4w9WgXcQ" />
```

## 示例：代码块

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
