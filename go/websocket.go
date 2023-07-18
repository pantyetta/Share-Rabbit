package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sort"
	"sync"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

type Message struct {
	M_type string `json:"type"`
	Msg    string `json:"msg"`
}

type User struct {
	conn net.Conn
	uid  uint64
}

func (u *User) send(msg []byte, op ws.OpCode) {
	err := wsutil.WriteServerMessage(u.conn, op, msg)
	if err != nil {
		fmt.Println("write server err", err)
		return
	}
}

type Chat struct {
	mu    sync.RWMutex
	users []*User
	seq   uint64
}

func (c *Chat) send(u *User, msg []byte, op ws.OpCode) {
	for _, v := range c.users {
		if v.uid == u.uid {
			continue
		}
		v.send(msg, op)
	}
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

	without := make([]*User, len(c.users)-1)
	copy(without[:i], c.users[:i])
	copy(without[i:], c.users[i+1:])
	c.users = without

	return true
}

var chat = Chat{}

func echo() {
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
						send_msg.Msg = "pong"

						send_json, err := json.Marshal(&send_msg)
						if err != nil {
							fmt.Println(err)
						}

						user.send(send_json, op)
					}
				case "echo":
					{
						chat.send(user, msg, op)
					}
				}
			}
		}()
	}))
}
