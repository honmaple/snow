---
title: "插件 (Hooks)"
weight: 70
sort_by: "weight"
---

Snow 的插件系统通过 Hooks 机制在内容处理流程中插入自定义逻辑。

## 查看插件

```bash
snow hooks
# 输出: snakecase, assets(enabled), pelican, rewrite, filter, encrypt(enabled), shortcode(enabled), minify
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

`weight` 控制执行顺序，数值越小越先执行；相同 `weight` 时按插件名排序。默认启用：`assets`、`encrypt`、`shortcode`。

如果配置了 `hooks.<name>.enabled: true`，但该插件没有被注册，构建会直接报错。使用 `--debug` 构建或预览时，会输出实际挂载顺序，例如：

```text
Enabled hooks: assets(20), encrypt(50), shortcode(60)
```

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
| snakecase | ❌ | 模板上下文 snake_case 访问 |

> 当前构建流程会调用单项阶段插件方法，例如页面、栏目、模板集、输出 writer、构建前后处理。集合级接口（如 pages/sections/taxonomies 批量处理）保留给后续扩展，当前未挂载到构建流程。
