---
title: "links"
weight: 55
---

## Links

`links` 会把正文中的本地内容链接转换为目标 Page 或 Section 的最终输出路径。它默认启用，适合在内容中引用源文件，而不用关心目标内容配置了怎样的 `path`。

```markdown
[下一篇](./another.org)
[项目文档](@/docs/project.md)
```

构建后，如果 `another.org` 或 `docs/project.md` 能在当前语言内容中找到，链接会改为对应页面或栏目的最终路径。

## 支持的链接

| 写法 | 说明 |
|------|------|
| `./another.md` | 相对当前内容文件所在目录 |
| `../intro.md` | 相对当前内容文件所在目录向上查找 |
| `@/docs/index.md` | 从 `content/` 根目录查找 |
| `./another.md#title` | 保留锚点 |
| `./another.md?from=home` | 保留查询参数 |

以下链接会保持原样：

- 以 `/` 开头的站点绝对路径
- `http://`、`https://`、`mailto:`、`tel:` 等协议链接
- 图片、附件、static 文件等非内容链接

未找到目标内容时，Snow 会输出 warning，并保留原链接。
