---
title: "snakecase"
weight: 70
---

## Snakecase

将模板上下文中的 Go 结构体字段和方法包装为小写下划线风格，便于在模板中使用更接近 Django/Jinja 的命名。

```yaml
hooks:
  snakecase:
    enabled: true
```

启用后，模板可以使用 snake_case 访问页面对象：

```html
{{ page.title }}
{{ page.front_matter.get_string("tags") }}
{{ get_page("posts/hello").permalink }}
```

函数或方法的返回值也会继续包装，因此链式调用仍可使用 snake_case。`time.Time` 等常用值类型会保持原值，以兼容 `date` 等过滤器。
