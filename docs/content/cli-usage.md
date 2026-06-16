---
title: "命令行"
weight: 15
---

Snow 提供 `init`、`server`、`build`、`hooks` 四个常用命令。

## init

创建一个新站点：

```bash
snow init myblog
```

如果不传目录名，Snow 会在当前目录初始化：

```bash
snow init
```

初始化过程会询问站点标题、作者、URL 前缀，以及是否创建第一篇文章。

## server

启动开发服务器：

```bash
snow server
```

常用参数：

```bash
snow server --listen 127.0.0.1:8088
snow server --autoload
snow server --root-dir myblog
snow server --config other.yaml
snow server --debug
snow server --mode publish
snow server --include-drafts
```

短参数：

```bash
snow server -l 127.0.0.1:8088
snow server -R
snow server -r myblog
snow server -c other.yaml
snow server -D
snow server -m publish
```

`--autoload` 会监听内容、模板、静态文件等变化，并触发重新构建与浏览器刷新。

## build

构建站点：

```bash
snow build
```

常用参数：

```bash
snow build --clean
snow build --root-dir myblog
snow build --output-dir dist
snow build --config other.yaml
snow build --dry-run
snow build --debug
snow build --mode publish
snow build --include-drafts
```

短参数：

```bash
snow build -C
snow build -r myblog
snow build -o dist
snow build -c other.yaml
snow build -D
snow build -m publish
```

`--dry-run` 会执行构建流程，但不会写入输出文件。`--clean` 会在构建前清理输出目录中非隐藏文件。

## hooks

查看已注册 Hook：

```bash
snow hooks
```

输出会标记当前配置中启用的 Hook，例如：

```text
snakecase, assets(enabled), pelican, rewrite, filter, encrypt(enabled), shortcode(enabled), minify
```

## 共享参数

`server` 和 `build` 都支持：

| 参数               | 短参数 | 说明                  |
|--------------------|--------|-----------------------|
| `--config`         | `-c`   | 指定配置文件          |
| `--debug`          | `-D`   | 启用调试模式          |
| `--root-dir`       | `-r`   | 指定站点根目录        |
| `--mode`           | `-m`   | 使用配置中的指定 mode |
| `--include-drafts` | -      | 包含草稿内容          |

`build` 额外支持：

| 参数           | 短参数 | 说明                       |
|----------------|--------|----------------------------|
| `--output-dir` | `-o`   | 覆盖输出目录               |
| `--clean`      | `-C`   | 构建前清理输出目录         |
| `--dry-run`    | -      | 只执行构建流程，不写入文件 |

`server` 额外支持：

| 参数         | 短参数 | 说明                   |
|--------------|--------|------------------------|
| `--listen`   | `-l`   | 指定监听地址           |
| `--autoload` | `-R`   | 监听文件变化并自动刷新 |

## root-dir

`--root-dir` 会在执行命令前切换到指定目录，因此配置文件、内容目录、模板目录、静态目录等相对路径都会以该目录为根。

```bash
snow build --root-dir myblog
snow server --root-dir myblog --autoload
```
