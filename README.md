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

## 项目结构

### /cmd/server

服务器二进制命令行启动代码

### /pkg/option

系统配置，外部服务资源初始化

### /pkg/presentaion

表现层，http、grpc、定时任务、队列consumer等

### /pkg/user

业务模块，所有的业务模块内部都按照Clean Architecture / DDD的方式组织代码

每个业务模块可以想象为代码级别的微服务

#### /pkg/user/app

业务逻辑层

表现层所有的请求都由user/app/handler内封装的各种handler来处理

多个handler之间共用的逻辑被封装为相应的user/app/service，service只应该被handler调用，不应该被直接使用

#### /pkg/user/domain

领域逻辑

#### /pkg/user/infra

基础设施，实现/user/app/adapter内约定的接口行为

### /pkg/utils

各种不包含具体业务逻辑的工具函数、对象、接口

代码不应该直接放到utils package下面，应该被放到utils内的子package内，避免utils成为一个什么都能放的筐
