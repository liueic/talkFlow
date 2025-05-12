package api

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"

	"talkFlow/config"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Client struct {
	conn   *websocket.Conn
	roomID string
	userID string
	send   chan []byte
	once   sync.Once // 新增
}

type RoomHub struct {
	rooms map[string]map[string]*Client
	lock  sync.Mutex
}

var Hub = RoomHub{
	rooms: make(map[string]map[string]*Client),
}

// 创建 WebSocket 连接
func TalkHandler(c *gin.Context) {
	roomID := c.Query("join_code")
	userID := c.Query("id")

	// 校验房间是否有效
	roomCollection := config.DB.Collection("rooms")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var room struct {
		ExpireTime time.Time `bson:"expire_time"`
		Status     int       `bson:"status"`
	}
	err := roomCollection.FindOne(ctx, bson.M{"join_code": roomID}).Decode(&room)
	if err != nil || room.Status != 0 || room.ExpireTime.Before(time.Now()) {
		c.JSON(400, gin.H{"error": "房间不存在或已过期"})
		return
	}

	// 升级为 WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := &Client{
		conn:   conn,
		roomID: roomID,
		userID: userID,
		send:   make(chan []byte, 256),
	}

	// 添加到房间
	Hub.lock.Lock()
	if Hub.rooms[roomID] == nil {
		Hub.rooms[roomID] = make(map[string]*Client)
	}
	Hub.rooms[roomID][userID] = client
	Hub.lock.Unlock()

	go client.readPump()
	go client.writePump()
}

// 读取消息并广播
func (c *Client) readPump() {
	defer c.cleanup()

	c.conn.SetReadDeadline(time.Now().Add(5 * time.Minute))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(5 * time.Minute))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			println("readPump exit:", err.Error())
			break
		}
		// 忽略心跳
		if string(message) == "ping" {
			continue
		}
		c.broadcastToRoom(message)
	}
}

// 写消息到客户端
func (c *Client) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.cleanup()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, nil)
				return
			}
			if err := c.conn.WriteMessage(websocket.BinaryMessage, msg); err != nil {
				println("writePump write error:", err.Error())
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// 广播消息到同房间用户
func (c *Client) broadcastToRoom(msg []byte) {
	Hub.lock.Lock()
	defer Hub.lock.Unlock()

	for uid, peer := range Hub.rooms[c.roomID] {
		if uid != c.userID {
			select {
			case peer.send <- msg:
			default:
				go peer.cleanup()
			}
		}
	}
}

// 清理连接资源
func (c *Client) cleanup() {
	c.once.Do(func() { // 保证只执行一次
		c.conn.Close()
		Hub.lock.Lock()
		defer Hub.lock.Unlock()

		if _, ok := Hub.rooms[c.roomID]; ok {
			delete(Hub.rooms[c.roomID], c.userID)
			if len(Hub.rooms[c.roomID]) == 0 {
				delete(Hub.rooms, c.roomID)
			}
		}
		close(c.send)
	})
}

// 定时清理已过期房间（main 启动时调用一次即可）
func StartRoomCleaner() {
	go func() {
		for {
			time.Sleep(time.Minute)
			roomCollection := config.DB.Collection("rooms")

			Hub.lock.Lock()
			for roomID, users := range Hub.rooms {
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				var room struct {
					ExpireTime time.Time `bson:"expire_time"`
					Status     int       `bson:"status"`
				}
				err := roomCollection.FindOne(ctx, bson.M{"join_code": roomID}).Decode(&room)
				cancel()

				if err != nil || room.Status != 0 || room.ExpireTime.Before(time.Now()) {
					for _, client := range users {
						client.conn.Close()
					}
					delete(Hub.rooms, roomID)
				}
			}
			Hub.lock.Unlock()
		}
	}()
}
