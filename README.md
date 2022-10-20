# simple-amqp-pipeline

### 功能说明
RabbitMQ消息队列迁移工具

### 使用方法
将配置文件拷贝至指定目录中
amqp-pipeline -path /root/conf.d


### 配置文件说明

| 字段           | 说明          |
|--------------|-------------|
| rabbitmq-src | 源MQ连接参数     |
| rabbitmq-dst | 目标MQ连接参数    |
| workers      | 源MQ并发读取数量   |
