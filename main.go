package main

import (
  "fmt"
  "log"
  "net/http"

  "github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
  ReadBufferSize:  1024,
  WriteBufferSize: 1024,
  CheckOrigin: func(r *http.Request) bool {
    return true
  },
}

var chat = &Chat{
  rooms:            make(map[string]Room),
  broadcastChannel: make(chan Message),
}

func main() {
  http.HandleFunc("/ws", serveWs)

  go chat.HandleBroadcast()

  err := http.ListenAndServe(":8080", nil)
  if err != nil {
    log.Fatal("ListenAndServe: ", err)
  }
}

func serveWs(w http.ResponseWriter, r *http.Request) {
  conn, err := upgrader.Upgrade(w, r, nil)
  if err != nil {
    log.Println(err)
    return
  }
  defer conn.Close()

  for {
    var message Message
    err := conn.ReadJSON(&message)
    if err != nil {
      log.Println(err)
      return
    }

    switch message.Action {
    case CreateRoomAction:
      chat.CreateRoom(message.RoomName)
      m := fmt.Sprintf("created room with name %s", message.RoomName)
      if err := sendResponseMessage(conn, m); err != nil {
        return
      }

    case JoinRoomAction:
      room, found := chat.FindRoom(message.RoomName)
      if !found {
        m := fmt.Sprintf("room `%s` not found", message.RoomName)
        if err := sendResponseMessage(conn, m); err != nil {
          return
        }
        continue
      }

      room.Join(conn)
      if err := sendResponseMessage(conn, "Joined room:"+message.RoomName); err != nil {
        return
      }
      log.Println("Joined room:", room.Name)

    case LeaveRoomAction:
      room, found := chat.FindRoom(message.RoomName)
      if !found {
        m := fmt.Sprintf("room `%s` not found", message.RoomName)
        if err := sendResponseMessage(conn, m); err != nil {
          return
        }
        continue
      }

      room.Leave(conn)
      if err := sendResponseMessage(conn, "Left room:"+message.RoomName); err != nil {
        return
      }
      log.Println("Left room:", room.Name)

    case SendMessageAction:
      chat.BroadcastMessage(message, conn)

    default:
      resMessage := fmt.Sprintf("unknown action: %s", message.Action)
      if err := conn.WriteMessage(websocket.TextMessage, []byte(resMessage)); err != nil {
        log.Println(err)
        return
      }
    }
  }
}

func sendResponseMessage(conn *websocket.Conn, message string) error {
  if err := conn.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
    log.Println(err)
    return err
  }

  return nil
}
