# Sagas 例子

### 运行TC
> go run tranman.go

### 运行Mysql
> docker run -itd  -v $PATH_TO_SAGAS/sagas/examples/sql/:/docker-entrypoint-initdb.d/ -p3306:3306 -e MYSQL_ROOT_PASSWORD=123456 mysql:latest

### 运行Redis
> docker run -itd -p6379:6379 redis

### 环境变量
> export SAGA_MYSQL_USERNAME=root
> export SAGA_MYSQL_PASSWORD=123456
> export SAGA_MYSQL_MASTER_HOST=127.0.0.1
> export SAGA_MYSQL_SLAVE_HOST=127.0.0.1
> export SAGA_MYSQL_PORT=3306




例子 | 场景 | 现象 | 严重程度|测试方法|预期|处理方案|进度
---|---|---|---|---|---|---|---|
example1|正常事物提交 |事物完整实现|无|启动4个服务|txgroup.success = 1,产生3条补偿数据|无|无|完成
example2|其中一个业务异常|业务异常|高|启动A,B 2个服务，B服务返回业务处理失败|txgroup.success == 3|需要TC补偿|完成
example3|saga网络异常|分布式事物基础设施异常|高|||需要TC补偿
example4|业务补偿接口失败|业务异常|中|||业务处理，TC告警





