package main

import (
  "github.com/gorilla/websocket"
  "sync"
)

type Room struct {
  Name         string
  Participants map[*websocket.Conn]bool
  mutex        sync.Mutex
}

func (r *Room) Join(conn *websocket.Conn) {
  r.mutex.Lock()
  defer r.mutex.Unlock()

  r.Participants[conn] = true
}

func (r *Room) Leave(conn *websocket.Conn) {
  r.mutex.Lock()
  defer r.mutex.Unlock()

  delete(r.Participants, conn)
}
