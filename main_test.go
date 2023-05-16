package main

import (
  "github.com/gorilla/websocket"
  "github.com/stretchr/testify/assert"
  "net/http"
  "net/http/httptest"
  "strings"
  "testing"
)

func TestHandleWebSocket(t *testing.T) {
  go chat.HandleBroadcast()

  // create server
  server := httptest.NewServer(http.HandlerFunc(serveWs))
  wsUrl := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"

  // create client 1
  wsConn1, _, err := websocket.DefaultDialer.Dial(wsUrl, nil)
  assert.NoError(t, err)

  // create client 2
  wsConn2, _, err := websocket.DefaultDialer.Dial(wsUrl, nil)
  assert.NoError(t, err)

  defer func() {
    server.Close()
    wsConn1.Close()
    wsConn2.Close()
  }()

  // send message to create room
  err = wsConn1.WriteJSON(Message{
    Action:   CreateRoomAction,
    RoomName: "test_room",
  })
  assert.NoError(t, err)

  // read response message
  var responseMessage ResponseMessage
  err = wsConn1.ReadJSON(&responseMessage)
  assert.NoError(t, err)

  expectedResponseMessage := ResponseMessage{
    Content: "created room with name test_room",
    Type:    Info,
  }
  assert.Equal(t, expectedResponseMessage, responseMessage)

  // send message to join client 1 to room
  err = wsConn1.WriteJSON(Message{
    Action:   JoinRoomAction,
    RoomName: "test_room",
  })
  assert.NoError(t, err)

  // read response message
  err = wsConn1.ReadJSON(&responseMessage)
  assert.NoError(t, err)

  expectedResponseMessage = ResponseMessage{
    Content: "Joined room:test_room",
    Type:    Info,
  }
  assert.Equal(t, expectedResponseMessage, responseMessage)

  // send message to join client 2 to room
  err = wsConn2.WriteJSON(Message{
    Action:   JoinRoomAction,
    RoomName: "test_room",
  })
  assert.NoError(t, err)

  // read response message
  err = wsConn2.ReadJSON(&responseMessage)
  assert.NoError(t, err)

  expectedResponseMessage = ResponseMessage{
    Content: "Joined room:test_room",
    Type:    Info,
  }
  assert.Equal(t, expectedResponseMessage, responseMessage)

  // send message to send message
  err = wsConn1.WriteJSON(Message{
    RoomName: "test_room",
    Action:   SendMessageAction,
    Content:  "Hello World!",
  })
  assert.NoError(t, err)

  // read response message
  err = wsConn2.ReadJSON(&responseMessage)
  assert.NoError(t, err)

  expectedResponseMessage = ResponseMessage{
    Content:  "Hello World!",
    Type:     Text,
    RoomName: "test_room",
  }
  assert.Equal(t, expectedResponseMessage, responseMessage)
}
