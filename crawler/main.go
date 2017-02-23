package main

// run RabbitMQ server in the console: $ rabbitmq-server

import (
  "fmt"
  "net/http"
  "time"
  "log"

  "github.com/PuerkitoBio/gocrawl"
  "github.com/PuerkitoBio/goquery"
  "github.com/streadway/amqp"
)

type Ext struct {
  *gocrawl.DefaultExtender
  PageChannel chan Page
}

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

// crawl methods
func (e *Ext) Visit(ctx *gocrawl.URLContext, res *http.Response, doc *goquery.Document) (interface{}, bool) {
  fmt.Printf("Visit: %s\n", ctx.URL())
  url := fmt.Sprintf("%s", ctx.URL())
  body := fmt.Sprintf("%s", res.Body)
  e.PageChannel <- Page{url, body}
  return nil, true
}

func (e *Ext) Filter(ctx *gocrawl.URLContext, isVisited bool) bool {
  if isVisited {
    return false
  }
  if ctx.URL().Host == "github.com" || ctx.URL().Host == "golang.org" || ctx.URL().Host == "0value.com" {
    return true
  }
  return false
}

func main() {
  pagechan := make(chan Page)

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
    false,   // delete when unused
    false,   // exclusive
    false,   // no-wait
    nil,     // arguments
  )
  failOnError(err, "Failed to declare a queue")

  //crawl
  ext := &Ext{&gocrawl.DefaultExtender{}, pagechan}
  // Set custom crawl options
  opts := gocrawl.NewOptions(ext)
  opts.CrawlDelay = 1 * time.Second
  opts.LogFlags = gocrawl.LogError
  opts.SameHostOnly = false
  opts.MaxVisits = 100

  c := gocrawl.NewCrawlerWithOptions(opts)
  go c.Run("http://0value.com")

  for page := range pagechan {
    // push page to RabbitMQ
    err = ch.Publish(
      "",     // exchange
      q.Name, // routing key
      false,  // mandatory
      false,  // immediate
      amqp.Publishing {
        ContentType: "text/plain",
        Body:        []byte(page.Url),
      })
    log.Printf(" [x] Sent %s", page.Url)
    failOnError(err, "Failed to publish a message")
  }
}
