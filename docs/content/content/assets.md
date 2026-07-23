---
title: "附件资源 (Assets)"
weight: 25
---

Page Bundle 和 Section 都可以声明附件资源。附件会从固定的 `content/` 目录中读取，并在构建时复制到对应页面或栏目的输出目录。

## Page Bundle 附件

包含 `index.{md,org,html}` 的目录视为 Page Bundle。目录内除 `index.*` 之外的文件都可以作为页面附件：

```
content/
└── posts/
    └── hello/
        ├── index.md
        ├── image.png
        └── gallery/
            └── cover.png
```

如果页面没有配置 `assets`，Snow 会收集 Page Bundle 内的全部附件。

```yaml
---
title: "Hello"
---
```

如果配置了 `assets`，则只收集匹配到的附件：

```yaml
---
title: "Hello"
assets:
  - image.png
  - gallery/**/*.png
---
```

`assets` 支持 glob 匹配，匹配规则使用 [doublestar](https://github.com/bmatcuk/doublestar)。

## Section 附件

Section 附件在 `_index.{md,org,html}` 的 FrontMatter 中声明：

```yaml
---
title: "Blog"
assets:
  - cover.png
  - media/**/*.png
---
```

Section 不会默认收集目录内所有文件；只有 `assets` 中匹配到的文件会被复制。

## 路径规则

`assets` 中的路径必须是相对当前 Page Bundle 或 Section 目录的干净路径：

- 不能是绝对路径
- 不能以 `./` 开头
- 不能包含 `../`
- 不能包含会被 `path.Clean` 改写的路径片段

合法示例：

```yaml
assets:
  - image.png
  - media/cover.png
  - gallery/**/*.jpg
```

非法示例：

```yaml
assets:
  - /image.png
  - ./image.png
  - ../image.png
  - media/../image.png
```

## 输出位置

附件会根据 Page 或 Section 最终的 `Path` 输出。

如果 `Path` 以 `/` 结尾，附件会输出到该目录下：

| 内容路径 | 输出路径 | 附件 | 附件输出 |
| --- | --- | --- | --- |
| `posts/hello/index.md` | `/posts/hello/` | `image.png` | `/posts/hello/image.png` |
| `posts/hello/index.md` | `/posts/hello/` | `gallery/cover.png` | `/posts/hello/gallery/cover.png` |

如果 `Path` 是具体 HTML 文件，附件会输出到该 HTML 文件所在目录：

| 内容路径 | 输出路径 | 附件 | 附件输出 |
| --- | --- | --- | --- |
| `posts/hello/index.md` | `/posts/hello.html` | `image.png` | `/posts/image.png` |
| `posts/hello/index.md` | `/posts/hello.html` | `gallery/cover.png` | `/posts/gallery/cover.png` |
| `blog/_index.md` | `/blog.html` | `cover.png` | `/cover.png` |
| `blog/_index.md` | `/blog.html` | `media/cover.png` | `/media/cover.png` |

子目录结构会保留，只有 Page 或 Section 的输出基准目录会变化。

## 模板变量

附件会挂载到 `page.Assets` 或 `section.Assets`：

| 字段 | 说明 |
| --- | --- |
| `File` | 附件源文件信息 |
| `Path` | 附件输出相对 URL |
| `Permalink` | 附件输出绝对 URL |

示例：

```html
{% for asset in page.Assets %}
  <a href="{{ asset.Permalink }}">{{ asset.File.Name }}</a>
{% endfor %}
```
