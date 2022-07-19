# 专业技能考核

实现基本的账号登录、注册接口

## 运行需求

- docker
- golang
- gow (go install github.com/mitranim/gow@latest)
- wire (go install github.com/google/wire/cmd/wire@latest)

## Make命令

- `make serve` 启动本地测试服务
- `make clean` 清除docker容器
- `make test` 执行单元测试
- `make alltest` 执行所有测试(单元测试和数据库集成测试)
- `make wire` 依赖注入代码更新

`make alltest`需要初始化完成的数据库，可以用`make serve`来实现初始化，只需要初始化一次即可，但在`make clean`之后需要重新初始化

## 接口测试

VSCode可以使用`test/api.http`测试脚本对本地启动好的服务进行测试，需要安装`REST Client`插件
