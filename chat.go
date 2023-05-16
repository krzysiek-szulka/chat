package main

import (
  "github.com/gorilla/websocket"
  "log"
  "sync"
)

type Chat struct {
  rooms            map[string]Room
  mutex            sync.Mutex
  broadcastChannel chan Message
}

type ActionType string

const (
  CreateRoomAction  ActionType = "create_room"
  JoinRoomAction    ActionType = "join_room"
  LeaveRoomAction   ActionType = "leave_room"
  SendMessageAction ActionType = "send_message"
)

type Message struct {
  RoomName string     `json:"room_name"`
  Action   ActionType `json:"action"`
  Content  string     `json:"content"`
}

func (c *Chat) CreateRoom(name string) {
  c.mutex.Lock()
  defer c.mutex.Unlock()

  _, found := c.rooms[name]
  if found {
    log.Printf("room `%s` already exists", name)
    return
  }

  c.rooms[name] = Room{
    Name:         name,
    Participants: map[*websocket.Conn]bool{},
  }

  log.Printf("created room with name %s", name)
  log.Println("Existing rooms:", c.rooms)
}

func (c *Chat) FindRoom(name string) (*Room, bool) {
  room, found := c.rooms[name]
  return &room, found
}

func (c *Chat) BroadcastMessage(message Message, senderConnection *websocket.Conn) {
  // check if sender can send message to room
  room, found := c.FindRoom(message.RoomName)
  if !found {
    log.Printf("room `%s` not found", message.RoomName)
    return
  }

  if _, found := room.Participants[senderConnection]; !found {
    log.Printf("participant not found in room `%s`", message.RoomName)
    return
  }

  c.broadcastChannel <- message
}

func (c *Chat) HandleBroadcast() {
  for message := range c.broadcastChannel {
    room, found := c.FindRoom(message.RoomName)
    if !found {
      log.Printf("room `%s` not found", message.RoomName)
      continue
    }

    for conn := range room.Participants {
      if err := conn.WriteMessage(websocket.TextMessage, []byte(message.Content)); err != nil {
        log.Println(err)
        return
      }
    }
  }
}
