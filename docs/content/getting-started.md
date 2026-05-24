# 快速开始

## 安装

```bash
# 从源码安装
go install github.com/honmaple/snow@latest

# 或手动编译
git clone https://github.com/honmaple/snow --depth=1
cd snow
go mod tidy
go build .
```

## 创建站点

```bash
snow init [目录名]
```

交互式提示：

```
$ snow init myblog
Welcome to snow 0.1.6.
> Where do you want to create your new web site? [.] myblog
> What will be the title of this web site? [snow]
> Who will be the author of this web site? honmaple
> What is your URL prefix? (no trailing slash) [http://127.0.0.1:8000]
> Do you want to create first page? [Y/n]
```

初始化后生成：

```
myblog/
├── config.yaml
├── content/
│   └── posts/
│       └── first-page.md
├── static/
├── templates/
└── themes/
```

## 构建站点

```bash
# 基础构建
snow build

# 清理输出目录
snow build --clean
snow build -C

# 调试模式
snow build --debug
snow build -D

# 指定配置文件
snow build --config other.yaml
snow build -c other.yaml

# 指定根目录
snow build --root-dir .
snow build -r .

# 指定输出目录
snow build --output-dir dist
snow build -o dist

# 预演（不实际写入文件）
snow build --dry-run

# 指定模式
snow build --mode publish
snow build -m publish

# 包含草稿
snow build --include-drafts
```

## 开发服务器

```bash
# 启动
snow server

# 指定监听地址
snow server --listen 127.0.0.1:8088
snow server -l 127.0.0.1:8088

# 热重载（监听文件变化）
snow server --autoload
snow server -R

# 调试模式
snow server --debug
snow server -D

# 指定模式
snow server --mode publish
snow server -m publish

# 包含草稿
snow server --include-drafts
```

> `server` 与 `build` 共享 `--config`、`--debug`、`--mode`、`--include-drafts` 参数。默认监听地址为配置中的 `base_url`。

## 查看插件

```bash
snow hooks
# 输出: assets(enabled), encrypt(enabled), filter, minify, pelican(enabled), rewrite(enabled), shortcode(enabled)
```

## 下一步

- [目录结构](/directory-structure) — 站点文件组织
- [配置](/configuration) — 完整配置参考
- [内容管理](/content) — 页面、栏目、分类
