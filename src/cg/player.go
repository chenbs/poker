//对在线玩家的管理
package cg

import (
	"fmt"
)

type Player struct {
	Username string
	Password string
	Number   string
	IsReady  int

	mq chan *Message //等待收取的消息
}

func NewPlayer() *Player {
	m := make(chan *Message, 1024)
	player := &Player{"", "", "", 0, m}

	go func(p *Player) {
		for {
			msg := <-p.mq
			fmt.Println(p.Username, "received message:", msg.Content)
		}
	}(player)

	return player
}
