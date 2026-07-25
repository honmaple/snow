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

## 基础配置

`weight` 控制执行顺序，数值越小越先执行；相同 `weight` 时按插件名排序。默认启用：`assets`、`encrypt`、`links`、`shortcode`。

如果配置了 `hooks.<name>.enabled: true`，但该插件没有被注册，构建会直接报错。使用 `--debug` 构建或预览时，会输出实际挂载顺序，例如：

```text
Enabled hooks: assets(20), encrypt(50), links(55), shortcode(60)
```

## 内置插件

| 插件 | 默认启用 | 说明 |
|------|----------|------|
| [assets](assets/) | ✅ | 静态资源处理 |
| [encrypt](encrypt/) | ✅ | 内容加密 |
| [links](links/) | ✅ | 将正文中的本地内容链接转换为最终路径 |
| [shortcode](../content/shortcodes/) | ✅ | 内容中嵌入可复用组件 |
| [mount](mount/) | ❌ | 挂载外部文件或目录 |
| [alias](alias/) | ❌ | 旧 URL 重定向到新 URL |
| [pelican](pelican/) | ❌ | 文档格式转换 |
| [rewrite](rewrite/) | ❌ | FrontMatter 重写 |
| [filter](filter/) | ❌ | 页面筛选 |
| [minify](minify/) | ❌ | 输出压缩 |
| [snakecase](snakecase/) | ❌ | 模板上下文 snake_case 访问 |
