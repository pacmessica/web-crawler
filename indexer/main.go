package main

import (
  "fmt"
  "log"
  "encoding/json"
  "strconv"
  "strings"
  "regexp"

  "github.com/streadway/amqp"
  "gopkg.in/redis.v5"
  "golang.org/x/net/html"
)

type Page struct {
  Url string
  Body string
}

var (
  // may want to improve
  // possible matches include: "hello", "I'm", "house-boat", "-'---"
  // non-matches: "CSS3", "Je$$ica"
  wordsRegex = regexp.MustCompile(`[a-zA-Z\-']+`)
)

func failOnError(err error, msg string) {
  if err != nil {
    log.Fatalf("%s: %s", msg, err)
    panic(fmt.Sprintf("%s: %s", msg, err))
  }
}


func getWordsFromHtml(body string, ch chan string) (string){
  z := html.NewTokenizer(strings.NewReader(body))
  for {
    tokenType := z.Next()
    switch tokenType {
    case html.ErrorToken:
      // return z.Err()
      // fmt.Println("error", z.Err())
    case html.TextToken:
      token := z.Token()
      words := wordsRegex.FindAllSubmatch([]byte(token.Data), -1)
      for _, word := range words {
        ch <- strings.ToLower(string(word[0]))
      }
    }
  }
}

func saveWords(pageId int, data []byte, client *redis.Client) {
  var page Page
  json.Unmarshal([]byte(data), &page)
  ch := make(chan string)
  go getWordsFromHtml(page.Body, ch)
  for word := range ch {
    fmt.Println("word!", word)
    // saved in db as key: `word:${word}:pageIds` value: set of pageIds
    err := client.SAdd("word:"+word+":pageIds", strconv.Itoa(pageId)).Err()
    failOnError(err, "Failed to save word")
  }
}

func saveWebpage(pageId int, data []byte, client *redis.Client) {
  var page Page
  json.Unmarshal([]byte(data), &page)
  log.Printf("Saving page: %s", page.Url)
  pagedata := make(map[string]string)
  pagedata["url"] = page.Url
  pagedata["body"] = page.Body
  // save page to redis as `page:${pageId}`
  err := client.HMSet("page:"+strconv.Itoa(pageId), pagedata).Err()
  failOnError(err, "Failed to save page")
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
    go saveWebpage(pageId, data.Body, client)
    go saveWords(pageId, data.Body, client)
  }

  log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
  <-loop
}
