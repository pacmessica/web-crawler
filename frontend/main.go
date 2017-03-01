package main

import (
  "net/http"

  "github.com/pacmessica/frontend/handler"
)


func main() {
  // frontend server
  http.HandleFunc("/get-pages", handler.GetPages)
  http.Handle("/", http.FileServer(http.Dir("html")))
  err := http.ListenAndServe(":8080", nil)
  if err != nil {
    panic("Error: " + err.Error())
  }
}
