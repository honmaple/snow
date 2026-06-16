---
title: "快速开始"
weight: 10
---

## 安装
使用 **Homebrew**
```bash
brew install honmaple/tap/snow
```

从源码安装
```bash
go install github.com/honmaple/snow@latest
```

手动编译
```
git clone https://github.com/honmaple/snow --depth=1
cd snow
go mod tidy
go build .
```

## 创建站点

```bash
snow init [目录名]
```

示例：
```
$ snow init myblog
Welcome to snow 0.1.7.
> Where do you want to create your new web site? [.] myblog
> What will be the title of this web site? [snow]
> Who will be the author of this web site? honmaple
> What is your URL prefix? (no trailing slash) [http://example.com]
> Do you want to create first page? [Y/n]
```

初始化后生成：
```
myblog/
├── config.yaml
├── content/
│   └── posts/
│       └── hello-snow.md
```

## 预览站点

```bash
cd myblog
snow server --autoload
```

如果不想切换目录，也可以指定站点根目录：

```bash
snow server --root-dir myblog --autoload
```

## 构建站点

```bash
snow build
```

常见生产构建：

```bash
snow build --mode publish --clean 
```

更多命令和参数见 [命令行使用](/cli-usage)。

## 下一步

- [命令行使用](/cli-usage) — 完整 CLI 参数
- [目录结构](/directory-structure) — 站点文件组织
- [配置](/configuration) — 完整配置参考
- [内容管理](/content) — 页面、栏目、分类
