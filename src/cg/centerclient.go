package cg

import (
	"encoding/json"
	"errors"
	//"fmt"
	"ipc"
)

type CenterClient struct {
	*ipc.IpcClient
}

func (client *CenterClient) Login(player *Player) (ps []*Player, err error) {
	b, err := json.Marshal(*player)
	if err != nil {
		return
	}
	resp, err := client.Call("Login", string(b))
	if err == nil && resp.Code == "200" {
		if resp.Body == "" {
			return
		}
		err = json.Unmarshal([]byte(resp.Body), &ps)
		return
	}
	err = errors.New(resp.Code)
	return
}

func (client *CenterClient) Logout(name string) error {
	ret, _ := client.Call("Logout", name)
	if ret.Code == "200" {
		return nil
	}
	return errors.New(ret.Code)

}

func (client *CenterClient) LogoutDis(number string) error {
	ret, _ := client.Call("LogoutDis", number)
	if ret.Code == "200" {
		return nil
	}
	return errors.New(ret.Code)

}

func (client *CenterClient) Ready(name string) (state string, err error) {
	resp, err := client.Call("Ready", name)
	if err == nil && resp.Code == "200" {
		if resp.Body == "" {
			return
		}
		state = resp.Body
		return
	}
	return
}

/*func (client *CenterClient) ListPlayer(params string) (ps []*Player, err error) {
	resp, _ := client.Call("listplayer", params)
	if resp.Code != "200" {
		err = errors.New(resp.Code)
		return
	}
	err = json.Unmarshal([]byte(resp.Body), &ps)
	return
}
func (client *CenterClient) Broadcast(message string) error {
	m := &Message{Content: message} // 构造Message结构体
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}
	resp, _ := client.Call("broadcast", string(b))
	if resp.Code == "200" {
		return nil
	}
	return errors.New(resp.Code)
}
*/
