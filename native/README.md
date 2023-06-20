# operator的原生方法实现一个控制器

当svc的annotations含有以下字段，自动创建ingress
```yaml
···
annotations:
  ingress/http: true
···
```
