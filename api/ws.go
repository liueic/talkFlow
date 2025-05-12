package api

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"talkFlow/config"
	"talkFlow/models"
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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 查询房间信息到 room 结构体
	var room models.Room
	roomSQL := `SELECT id, creater, name, joiner, join_code, create_time, expire_time, status, ip FROM rooms WHERE join_code = ?`
	err := config.DB.QueryRowContext(ctx, roomSQL, roomID).Scan(
		&room.ID,
		&room.Creater,
		&room.Name,
		&room.JoinerStr, // 先用字符串接收
		&room.JoinCode,
		&room.CreateTime,
		&room.ExpireTime,
		&room.Status,
		&room.IP,
	)
	if err != nil {
		c.JSON(404, gin.H{
			"code":  40401,
			"error": "房间不存在",
		})
		log.Println("房间不存在:", roomID)
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

			Hub.lock.Lock()
			for roomID, users := range Hub.rooms {
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()

				// 查询房间的过期时间和状态
				var expireTime time.Time
				var status int
				query := `SELECT expire_time, status FROM rooms WHERE join_code = ?`
				err := config.DB.QueryRowContext(ctx, query, roomID).Scan(&expireTime, &status)

				// 如果查不到房间、房间已结束或已过期，则关闭所有连接并移除房间
				if err != nil || status != int(models.RoomOngoing) || expireTime.Before(time.Now()) {
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
