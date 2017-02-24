package main

import (
  "fmt"
  "log"
  "encoding/json"
  "strconv"
  "strings"

  "github.com/streadway/amqp"
  "gopkg.in/redis.v5"
  "golang.org/x/net/html"
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

func parseHtml(page string) (string){
  z := html.NewTokenizer(strings.NewReader(page))
  for {
    tokenType := z.Next()
    switch tokenType {
    case html.ErrorToken:
      // return z.Err()
      // fmt.Println("error", z.Err())
    case html.TextToken:
      token := z.Token()
      fmt.Println("foooo!", token.Data)
    }
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
    log.Printf("Received a message")
    pageId += 1
    // save page to redis as `webpage:${pageId}`
    go func(pageId int) {
      var page Page
      json.Unmarshal([]byte(data.Body), &page)
      log.Printf("Saving page: %s", page.Url)
      pagedata := make(map[string]string)
      pagedata["url"] = page.Url
      pagedata["body"] = page.Body
      err := client.HMSet("webpage:"+strconv.Itoa(pageId), pagedata).Err()
      failOnError(err, "Failed to save page")
      parseHtml(page.Body)
    }(pageId)
  }

  log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
  <-loop
}
