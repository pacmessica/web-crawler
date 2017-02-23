## Setup
1) Run RabbitMQ server in the console

```
$ rabbitmq-server
```
NOTE: must have RabbitMq installed before running. See https://www.rabbitmq.com/download.html

2) Install (if not installed) and run redis server
```
https://redis.io/download
```

3) `go get` the following packages:
```
github.com/streadway/amqp
gopkg.in/redis.v5
github.com/PuerkitoBio/gocrawl
github.com/PuerkitoBio/goquery
github.com/streadway/amqp
golang.org/x/net/html
```
