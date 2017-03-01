package main

import (
  "fmt"
  "net/http"
  "log"

  "github.com/micro/go-micro/client"
	"github.com/micro/go-micro/cmd"
	"github.com/micro/go-micro/metadata"
  pf "github.com/pacmessica/indexer/proto"
  "golang.org/x/net/context"
  "github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

func getPagesHandler(w http.ResponseWriter, r *http.Request) {
  conn, err := upgrader.Upgrade(w, r, nil)
  if err != nil {
    log.Println("upgrade:", err)
    return
  }
  defer conn.Close()

  for {
    messageType, message, err := conn.ReadMessage()
    if err != nil {
      log.Println("read:", err)
      return
    }
    log.Printf("recv: %s", message)
    fmt.Printf("foo message %T\n", message)
    response := GetPagesFromQuery(message)

    err = conn.WriteMessage(messageType, response);
    if  err != nil {
      log.Println("write:", err)
      return
    }
  }
}

func GetPagesFromQuery(query []byte) ([]byte){
  // register go-micro client
  cmd.Init()
  fmt.Printf("foo query %T\n", query)
  newstring := string(query)
  fmt.Printf("foo newstring %T\n", newstring)
  fmt.Println("[GetPagesFromQuery] Request: ", query)
	// Use the generated client stub
	cl := pf.NewPageGetterClient("pagegetter", client.DefaultClient)

	// Set arbitrary headers in context
	ctx := metadata.NewContext(context.Background(), map[string]string{
		"X-User-Id": "john",
		"X-From-Id": "script",
	})

  // Make request
  rsp, err := cl.GetPagesFromQuery(ctx, &pf.Request {
    Search: &pf.Search { Term: newstring },
  })
  if err != nil {
    fmt.Println(err)
    // return
  }
  fmt.Println(rsp.Pageids)
  fmt.Printf("foo Pageids %T\n", rsp.Pageids)
  if len(rsp.Pageids) < 1 {
    return []byte("No pages found")
  }
  return []byte(rsp.Pageids[0])
}

func main() {
  // frontend server
  http.HandleFunc("/get-pages", getPagesHandler)
  http.Handle("/", http.FileServer(http.Dir(".")))
  err := http.ListenAndServe(":8080", nil)
  if err != nil {
    panic("Error: " + err.Error())
  }
}
