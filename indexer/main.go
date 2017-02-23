package main

import (
  "fmt"
  "log"
  "encoding/json"
  "strconv"

  "github.com/streadway/amqp"
  "gopkg.in/redis.v5"
)

type Page struct {
  Url string
  Body string
}

func failOnError(err error, msg string) {
  if err != nil {
    log.Fatalf("%s: %s", msg, err)
    panic(fmt.Sprintf("%s: %s", msg, err))
  }
}

func main() {
  // connect to RabbitMQ
  conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
  failOnError(err, "Failed to connect to RabbitMQ")
  defer conn.Close()

  ch, err := conn.Channel()
  failOnError(err, "Failed to open a channel")
  defer ch.Close()

  q, err := ch.QueueDeclare(
    "webpage", // name
    false,   // durable
    false,   // delete when usused
    false,   // exclusive
    false,   // no-wait
    nil,     // arguments
  )
  failOnError(err, "Failed to declare a queue")

  // connect to redis
  client := redis.NewClient(&redis.Options{
    Addr:     "localhost:6379",
    Password: "", // no password set
    DB:       0,  // use default DB
  })

  pong, err := client.Ping().Result()
  fmt.Println(pong, err)

   //consume messages from RabbitMQ as msgs

  msgs, err := ch.Consume(
    q.Name, // queue
    "",     // consumer
    true,   // auto-ack
    false,  // exclusive
    false,  // no-local
    false,  // no-wait
    nil,    // args
  )
  failOnError(err, "Failed to register a consumer")

  loop := make(chan bool)

  var pageId int
  for data := range msgs {
    var page Page
    json.Unmarshal([]byte(data.Body), &page)
    log.Printf("Received a message: %s", page.Url)
    // save page to redis as `webpage:${pageId}`
    go func() {
      pageId += 1
      pagedata := make(map[string]string)
      pagedata["url"] = page.Url
      pagedata["body"] = page.Body
      err := client.HMSet("webpage:"+strconv.Itoa(pageId), pagedata).Err()
      failOnError(err, "Failed to save page")
    }()
  }

  log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
  <-loop
}
