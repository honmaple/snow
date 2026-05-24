# Snow 文档

Snow 是一个用 Go 语言编写的静态站点生成器，支持 Markdown、Org-mode、HTML 格式内容。

## 文档导航

- [快速开始](content/getting-started.md) — 安装、初始化、构建与开发服务器
- [目录结构](content/directory-structure.md) — 站点文件与目录布局
- [配置](content/configuration.md) — 完整配置选项参考
- [内容管理](content/content/) — 页面、栏目、分类、短代码、多语言、分页、输出格式
- [模板](content/templates/) — 模板语法、变量与函数
- [主题](content/themes.md) — 创建与使用主题
- [插件](content/hooks/) — 插件系统与内置插件

## 核心特性

- 多格式：Markdown (goldmark)、Org-mode、HTML
- 多语言：内置 i18n，目录/后缀/FrontMatter 三方式
- 灵活的 Taxonomy：标签、分类、作者及自定义归档
- 插件系统：短代码、加密、资源处理、压缩、内容重写
- 分页：Section 和 Taxonomy 级别
- 自定义输出：RSS、Atom、JSON 等
- 实时预览：WebSocket livereload
- 单二进制：编译为单文件，随处部署

## 项目链接

[GitHub](https://github.com/honmaple/snow) · [Issues](https://github.com/honmaple/snow/issues)
