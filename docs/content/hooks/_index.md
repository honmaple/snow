---
title: "插件 (Hooks)"
weight: 70
sort_by: "weight"
---

Snow 的插件系统通过 Hooks 机制在内容处理流程中插入自定义逻辑。

## 查看插件

```bash
snow hooks
# 输出: mount, snakecase, assets(enabled), pelican, rewrite, filter, encrypt(enabled), links(enabled), shortcode(enabled), minify, alias
```

## 配置

```yaml
hooks:
  assets:
    enabled: true
  mount:
    enabled: false
    option:
      - source: "/tmp/project-name/docs"
        target: "content/docs/project-name"
      - source: "/tmp/project-name/static/style.css"
        target: "static/style.css"
  encrypt:
    enabled: true
    weight: 2
    option:
      password: "123456"
  links:
    enabled: true
    weight: 55
  shortcode:
    enabled: true
    weight: 1
  minify:
    enabled: false
    option:
      html: true
      css: true
      js: true
  alias:
    enabled: false
    option:
      - "/old-url/:/new-url/"
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

`weight` 控制执行顺序，数值越小越先执行；相同 `weight` 时按插件名排序。默认启用：`assets`、`encrypt`、`links`、`shortcode`。

如果配置了 `hooks.<name>.enabled: true`，但该插件没有被注册，构建会直接报错。使用 `--debug` 构建或预览时，会输出实际挂载顺序，例如：

```text
Enabled hooks: assets(20), encrypt(50), links(55), shortcode(60)
```

## 内置插件

| 插件 | 默认启用 | 说明 |
|------|----------|------|
| shortcode | ✅ | 内容中嵌入可复用组件 |
| encrypt | ✅ | 内容加密 |
| links | ✅ | 将正文中的本地内容源文件链接转换为目标页面路径 |
| assets | ✅ | 静态资源处理 |
| mount | ❌ | 将本地文件或目录挂载到虚拟文件系统路径 |
| alias | ❌ | 旧 URL 重定向到新 URL |
| pelican | ❌ | 文档格式转换 |
| rewrite | ❌ | FrontMatter 重写 |
| minify | ❌ | 输出压缩 |
| filter | ❌ | 页面筛选 |
| snakecase | ❌ | 模板上下文 snake_case 访问 |

> 当前构建流程会调用实际挂载的阶段方法，例如页面、栏目、内容 store、模板集、输出 writer、构建前后处理。未被构建流程使用的集合级 hook 接口已移除。

`mount` hook 只影响构建读取使用的虚拟文件系统；开发服务器 watcher 不会监听被挂载的外部路径。

### mount

`mount` 可把本地文件或目录接入虚拟文件系统。构建流程仍然通过 `content`、`static`、`templates` 等虚拟路径读取文件，因此挂载后不需要让内容、静态文件或模板逻辑知道真实来源。

```yaml
hooks:
  mount:
    enabled: true
    option:
      - source: "/tmp/project-name/docs"
        target: "content/docs/project-name"
      - source: "/tmp/project-name/static/style.css"
        target: "static/style.css"
```

规则：

- `source` 是本地文件或目录路径。
- `target` 是虚拟文件系统路径，不能是空路径、绝对路径，也不能包含 `.`、`./` 或 `..` 这种需要再次清理的路径片段。
- 目录挂载后会把目录内容暴露到 `target` 下。
- 文件挂载后会把该文件暴露为 `target`。
- watcher 不监听 `source` 指向的外部路径；开发时修改外部挂载文件后，需要手动触发重建或重启预览。
