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
  "github.com/micro/go-micro"
  proto "github.com/pacmessica/indexer/proto"
  "golang.org/x/net/context"
)

type Page struct {
  Url string
  Body string
}

type PageGetter struct {
  client *redis.Client
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

func removeDuplicates(elements []string) []string {
  encountered := map[string]bool{}
  for v := range elements {
    encountered[elements[v]] = true
  }
  var result []string
  for key, _ := range encountered {
    result = append(result, key)
  }
  return result
}

func getPageIdsForQuery(query []string, client *redis.Client, ch chan []string) {
  ids, err := client.SInter(query...).Result()
  failOnError(err, "Failed while fetching pageIds")
  ch <- ids
}

func getPageIdsForQueries(queries [][]string, client *redis.Client) ([]string){
  log.Printf("[getPageIdsForQueries] Request: %s", queries)
  ch := make(chan []string)
  for _, query := range queries {
    for i, word := range query {
      query[i] = "word:"+word+":pageIds"
    }
    go getPageIdsForQuery(query, client, ch)
  }
  var ids []string
  for i:=0; i<len(queries); i++ {
    in := <- ch
    ids = append(ids, in...)
  }
  close(ch)
  return removeDuplicates(ids)
}

func (g *PageGetter) GetPagesFromQuery(ctx context.Context, req *proto.Request, rsp *proto.Result) error {
  log.Printf("[GetPagesFromQuery] Request: %s", req)
  queries := getQueries(req.Search)
  ids := getPageIdsForQueries(queries, g.client)
  log.Printf("[GetPagesFromQuery] Response: pageIds '%s'", ids)
  rsp.Pageids = ids
  return nil
}

func getQueries(req *proto.Search) ([][]string){
  var query []string
  var queries [][]string
  switch {
  case req.Term != "":
    query = append(query, req.Term)
    queries = append(queries, query)
  case req.And != nil:
    queries = getAndQueries(req.And)
  case req.Or != nil:
    queries = getOrQueries(req.Or)
  }
  return queries
}

func combineQueries(a [][]string, b []string) [][]string {
  // adds query 'b' to each query in 'a'
  combinedQuery := make([][]string, len(a))
  for i, _ := range a {
    combinedQuery[i] = append(a[i], b...)
  }
  return combinedQuery
}

func getAndQueries(req *proto.Search_And) ([][]string){
  andQueries := make([][]string, 1)
  for _, q := range req.Search {
    queries := getQueries(q)
    if len(queries) > 1 { // q looks like [[a][b][c]]
      // if addQueries = [[x y z]], we want [[x y z a] [x y z b] [x y z c]]
      temp := make([][]string, (len(queries)*len(andQueries)))
      index := 0
      for i, _ := range queries {
        c := combineQueries(andQueries, queries[i])
        for _, q := range c {
          temp[index] = append(temp[index], q...)
          index++
        }
      }
      andQueries = temp
    } else { // q looks like [[a b c]]
      for i:=0; i<len(andQueries); i++ {
        // add q to each query in addQueries
        // if addQueries = [[x y z]], we want [[x y z a b c]]
        andQueries[i] = append(andQueries[i], queries[0]...)
      }
    }
  }
  return andQueries
}

func getOrQueries(req *proto.Search_Or) ([][]string){
  var orQueries [][]string
  for _, q := range req.Search {
    queries := getQueries(q)
    for i:=0; i<len(queries); i++ {
      orQueries = append(orQueries, queries[i])
    }
  }
  return orQueries
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

  var pageId int
  go func() {
    for data := range msgs {
      log.Printf("Received a message")
      pageId += 1
      go saveWebpage(pageId, data.Body, client)
      go saveWords(pageId, data.Body, client)
    }
  }()
  log.Printf(" [*] Waiting for messages. To exit press CTRL+C")

  // initialize service
  service := micro.NewService(
    micro.Name("pagegetter"),
    micro.Version("latest"),
  )

  service.Init()

  proto.RegisterPageGetterHandler(service.Server(), &PageGetter{
    client: client,
  })

  if err := service.Run(); err != nil {
    log.Fatal(err)
  }
}
