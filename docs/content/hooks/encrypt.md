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
      password: "默认密码"
```

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
