package handler

import (
	"net/http"
  "fmt"
  "log"

  "github.com/micro/go-micro/client"
  "github.com/micro/go-micro/cmd"
	"github.com/micro/go-micro/metadata"
  pf "github.com/pacmessica/indexer/proto"
	"golang.org/x/net/context"
  "github.com/gorilla/websocket"
  "github.com/golang/protobuf/jsonpb"
)

var upgrader = websocket.Upgrader{}

func GetPages(w http.ResponseWriter, r *http.Request) {
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

    response := getPagesFromQuery(message)

    err = conn.WriteMessage(messageType, response);
    if  err != nil {
      log.Println("write:", err)
      return
    }
  }
}

func getPagesFromQuery(query []byte) ([]byte){
  var requestQuery pf.Request
  stringifiedQuery := string(query)
  err := jsonpb.UnmarshalString(stringifiedQuery, &requestQuery);
  if  err != nil {
    log.Println("[GetPagesFromQuery] Error: ", err)
  }
  // register go-micro client
  cmd.Init()
	// Use the generated client stub
	cl := pf.NewPageGetterClient("pagegetter", client.DefaultClient)

	// Set arbitrary headers in context
	ctx := metadata.NewContext(context.Background(), map[string]string{
		"X-User-Id": "john",
		"X-From-Id": "script",
	})

  // Make request
  rsp, err := cl.GetPagesFromQuery(ctx, &requestQuery)
  if err != nil {
    fmt.Println(err)
    // return
  }
  log.Printf("[GetPagesFromQuery] Response", rsp.Pageids)
  if len(rsp.Pageids) < 1 {
    return []byte("No pages found")
  }
  return []byte(rsp.Pageids[0])
}
