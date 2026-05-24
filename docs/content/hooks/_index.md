---
title: "插件 (Hooks)"
---

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
