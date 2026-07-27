---
title: "encrypt"
weight: 30
---

## Encrypt

内容加密。

```yaml
hooks:
  encrypt:
    enabled: true
    weight: 2
    option:
      mode: "cbc"
      password: "默认密码"
```

`mode` 支持 `cbc`、`ctr`、`cfb`、`ofb`，默认 `cbc`。

### 全篇加密

```yaml
---
password: "123456"
---
```

### 局部加密

```html
<shortcode encrypt password="123456">
加密内容
</shortcode>
```

### 模板中

```html
{{ page.Content | encrypt:"123456" }}
```
