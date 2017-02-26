package main

import (
  "fmt"

  "github.com/micro/go-micro/client"
	"github.com/micro/go-micro/cmd"
	"github.com/micro/go-micro/metadata"
  pf "github.com/pacmessica/indexer/proto"
  "golang.org/x/net/context"
)

func main() {
  cmd.Init()

	// Use the generated client stub
	cl := pf.NewPageGetterClient("pagegetter", client.DefaultClient)

	// Set arbitrary headers in context
	ctx := metadata.NewContext(context.Background(), map[string]string{
		"X-User-Id": "john",
		"X-From-Id": "script",
	})

	// Make request
	rsp, err := cl.GetPagesFromQuery(ctx, &pf.Request{
		Query: "this is a query",
	})
  if err != nil {
    fmt.Println(err)
    return
  }

  fmt.Println(rsp.Pageids)
}
