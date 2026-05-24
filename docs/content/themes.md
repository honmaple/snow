# 主题

Snow 通过主题系统管理站点外观。主题包含模板、静态资源和配置文件。

## 使用主题

```yaml
theme: "mytheme"
theme_dir: "themes"
```

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| `theme` | — | 主题名称，对应 `themes/{name}/` 目录 |
| `theme_dir` | `themes` | 主题存放目录 |

## 主题目录结构

```
themes/mytheme/
├── theme.yaml            # 主题配置（可选）
├── templates/            # 模板（固定名）
│   ├── index.html        # 首页
│   ├── post.html         # 文章页
│   ├── page.html         # 单页
│   ├── section.html      # 栏目页
│   ├── taxonomy.html     # 分类列表
│   ├── taxonomy.terms.html  # 分类详情
│   ├── partials/         # 局部模板
│   │   ├── header.html
│   │   ├── footer.html
│   │   ├── rss.xml
│   │   └── atom.xml
│   └── shortcodes/       # 短代码模板
│       └── youtube.html
├── static/               # 静态资源（固定名）
│   ├── css/
│   ├── js/
│   └── images/
└── i18n/                 # 翻译文件
    ├── en.yaml
    └── zh.yaml
```

`templates`、`static`、`i18n` 目录名不可修改。

## 主题配置 (theme.yaml)

主题根目录下的 `theme.yaml`（或 `.toml`、`.json`）会在加载时自动合并到站点配置：

```yaml
# themes/mytheme/theme.yaml
name: "MyTheme"
version: "1.0.0"
description: "A minimal theme"
author: "honmaple"

# 以下会合并到站点配置，但不覆盖站点已设置的项
params:
  mytheme_primary_color: "#333"
```

合并规则：站点已有配置项不会被主题配置覆盖。

## 模板查找优先级

`templates/` 目录的合并顺序（先匹配的文件优先）：

1. 站点 `templates/` 目录
2. 主题 `templates/` 目录
3. 内置默认主题 `templates/` 目录

因此只需在站点 `templates/` 中创建与主题同名的文件即可覆盖：

```
mysite/
├── templates/
│   └── post.html         # 覆盖主题的 post.html
└── themes/
    └── mytheme/
        └── templates/
            ├── index.html
            └── post.html   # 被覆盖
```

## 静态资源优先级

1. 站点 `static/` 目录
2. 主题 `static/` 目录

同名文件以站点为准。

## 创建最小主题

`themes/simple/`：

```
simple/
├── theme.yaml
├── templates/
│   ├── index.html
│   └── post.html
└── static/
    └── css/
        └── style.css
```

### index.html

```html
<!DOCTYPE html>
<html lang="{{ config.language }}">
<head>
  <meta charset="utf-8">
  <title>{{ config.title }}</title>
  <link rel="stylesheet" href="{{ config.base_url }}/css/style.css">
</head>
<body>
  <h1>{{ config.title }}</h1>
  {% for section in sections %}
  <section>
    <h2><a href="{{ section.Path }}">{{ section.Title }}</a></h2>
    <ul>
      {% for page in section.Pages %}
      <li><a href="{{ page.Path }}">{{ page.Title }}</a></li>
      {% endfor %}
    </ul>
  </section>
  {% endfor %}
</body>
</html>
```

### post.html

```html
<!DOCTYPE html>
<html lang="{{ page.Lang }}">
<head>
  <meta charset="utf-8">
  <title>{{ page.Title }} - {{ config.title }}</title>
  <link rel="stylesheet" href="{{ config.base_url }}/css/style.css">
</head>
<body>
  <article>
    <h1>{{ page.Title }}</h1>
    <time>{{ page.Date | date:"2006-01-02" }}</time>
    <div>{{ page.Content | safe }}</div>
  </article>
</body>
</html>
```
