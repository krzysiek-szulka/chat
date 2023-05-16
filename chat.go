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

type ResponseType string

const (
  Text ResponseType = "text"
  Info ResponseType = "info"
)

type ResponseMessage struct {
  RoomName string       `json:"room_name"`
  Type     ResponseType `json:"type"`
  Content  string       `json:"content"`
}

func (c *Chat) CreateRoom(name string) {
  _, found := c.rooms[name]
  if found {
    log.Printf("room `%s` already exists", name)
    return
  }

  // c.mutex.Lock() - it might be helpful if we add the user who created the room to the room immediately
  c.rooms[name] = Room{
    Name:         name,
    Participants: map[*websocket.Conn]bool{},
  }
  // c.mutex.Unlock()

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
      m := ResponseMessage{
        RoomName: message.RoomName,
        Type:     Text,
        Content:  message.Content,
      }
      if err := conn.WriteJSON(m); err != nil {
        log.Println(err)
        return
      }
    }
  }
}
