package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

type Message struct {
	M_type  string            `json:"type"`
	Msg     string            `json:"msg"`
	Tell    map[string]string `json:"tell"`
	History map[string]string `json:"history"`
}

type User struct {
	conn net.Conn
	uid  uint64
	name string
}

func (u *User) send(msg []byte, op ws.OpCode) {
	err := wsutil.WriteServerMessage(u.conn, op, msg)
	if err != nil {
		fmt.Println("write server err", err)
		return
	}
}

func (u *User) sendErr(msg string, op ws.OpCode) {
	var err_msg Message
	err_msg.M_type = "err"
	err_msg.Msg = msg
	err_json, err := json.Marshal(&err_msg)
	if err != nil {
		fmt.Println(err)
	}
	u.send(err_json, op)
}

func (u *User) rename(name string) {
	if name == "" {
		return
	}

	u.name = name
}

type Chat struct {
	mu    sync.RWMutex
	users []*User
	seq   uint64
}

func (c *Chat) Register(conn net.Conn) *User {

	user := &User{
		conn: conn,
	}

	c.mu.Lock()
	{
		user.uid = c.seq
		c.users = append(c.users, user)
		c.seq++
	}
	c.mu.Unlock()

	fmt.Println("new Client", user.uid)
	fmt.Println(c.users)
	return user
}

func (c *Chat) remove(u *User) bool {
	fmt.Println("remove", u.uid)
	if len(c.users) < 1 {
		return false
	}

	i := sort.Search(len(c.users), func(i int) bool {
		return c.users[i].uid >= u.uid
	})

	if i == len(c.users) && c.users[len(c.users)-1].uid != u.uid {
		return false
	}

	c.mu.Lock()
	without := make([]*User, len(c.users)-1)
	copy(without[:i], c.users[:i])
	copy(without[i:], c.users[i+1:])
	c.users = without
	c.mu.Unlock()

	return true
}

// u: send user, msg: message, op: mode
func (c *Chat) Broadcast(u *User, msg []byte, op ws.OpCode) {
	for _, v := range c.users {
		if v.uid == u.uid {
			continue
		}
		v.send(msg, op)
	}
}

func (c *Chat) Multicast(u *User, msg []byte, op ws.OpCode) {
	if u.name == "" {
		return
	}

	for _, v := range c.users {
		if v.name != u.name {
			continue
		}
		if v.uid == u.uid {
			continue
		}
		v.send(msg, op)
	}
}

var chat = Chat{}

func Websocket() {
	http.ListenAndServe(":80", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, _, _, err := ws.UpgradeHTTP(r, w)
		if err != nil {
			fmt.Println("ws.upgrade err", err)
		}

		user := chat.Register(conn)

		go func() {
			defer chat.remove(user)
			defer conn.Close()

			// 受け取り無限ループ
			for {
				var recive_msg Message
				var send_msg Message
				msg, op, err := wsutil.ReadClientData(conn)
				if err != nil {
					fmt.Println("read client err", err)
					return
				}

				if err := json.Unmarshal(msg, &recive_msg); err != nil {
					fmt.Println("json Unmarshal err", err)
				}
				fmt.Printf("msg %+v\n", recive_msg)

				switch recive_msg.M_type {
				case "ping":
					{
						send_msg.M_type = "ping"
						send_msg.Msg = "pong to " + user.name + "."

						send_json, err := json.Marshal(&send_msg)
						if err != nil {
							fmt.Println(err)
						}

						user.send(send_json, op)
					}
				case "echo":
					{
						chat.Broadcast(user, msg, op)
					}
				case "tell":
					{
						if user.name == "" {
							user.sendErr("need Name", op)
							break
						}
						key := strconv.FormatUint(user.uid, 10) + ":" + user.name + ":" + strconv.FormatInt(time.Now().UnixNano(), 10)
						if err := Add(key, recive_msg.Msg); err != nil {
							fmt.Println(err)
						}

						send_msg.M_type = "tell"
						send_msg.Tell = map[string]string{key: recive_msg.Msg}

						send_json, err := json.Marshal(&send_msg)
						if err != nil {
							fmt.Println(err)
						}
						chat.Multicast(user, send_json, op)
					}
				case "rename":
					{
						user.rename(recive_msg.Msg)
					}
				case "get":
					{
						if user.name == "" {
							user.sendErr("need Name", op)
							break
						}
						pattern := "*:" + regexp.QuoteMeta(user.name) + ":*"
						keys, _ := Keys(pattern)

						values := make(map[string]string)

						for _, v := range keys {
							value, _ := Get(v)
							values[v] = value
						}

						fmt.Println(keys)
						send_msg.M_type = "get"
						send_msg.History = values

						send_json, err := json.Marshal(&send_msg)
						if err != nil {
							fmt.Println(err)
						}

						user.send(send_json, op)
					}
				}
			}
		}()
	}))
}
