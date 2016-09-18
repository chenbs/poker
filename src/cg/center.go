//中央服务器
package cg

import (
	"encoding/json"
	"errors"
	"fmt"
	//"github.com/bitly/go-simplejson"
	"github.com/bradfitz/gomemcache/memcache"
	"ipc"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"sync"
)

type User struct {
	UserId       bson.ObjectId `bson:"_id,omitempty"`                      // 必须要设置bson:"_id" 不然mgo不会认为是主键
	UserName     string        `bson:"user_name" json:"user_name"`         //用户注册姓名
	UserPassword string        `bson:"user_password" json:"user_password"` //用户注册密码
	//UserRegion   string `bson:"user_region" json:"user_region"`
}

var _ ipc.Server = &CenterServer{} //确认实现了Server接口

var once sync.Once

type Message struct {
	From    string "from"
	To      string "to"
	Content string "content"
}

type CenterServer struct {
	servers map[string]ipc.Server
	players []*Player
	mutex   sync.RWMutex
	//rooms   []*Room
}

//var UserInfo map[string]*net.TCPConn

const (
	Url = "192.168.100.58:27017"
)

func NewCenterServer() *CenterServer {
	servers := make(map[string]ipc.Server)
	players := make([]*Player, 0)

	return &CenterServer{servers: servers, players: players}
}

//登录
func (server *CenterServer) Login(params string) (players string, err error) {

	session, err := mgo.Dial(Url)
	if err != nil {
		return
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)
	db := session.DB("game")
	collection := db.C("user")
	once.Do(onces)

	con := memcache.New("192.168.100.58:11211")
	if con == nil {
		fmt.Println("Failed to connect Memcache")
	}

	player := NewPlayer()
	err = json.Unmarshal([]byte(params), &player)
	if err != nil {
		return
	}

	//server.mutex.Lock()
	//defer server.mutex.Unlock()

	//重复登录检查
	for _, v := range server.players {
		if v.Username == player.Username {
			fmt.Println("The user has joined.")
			err = errors.New("The user has joined.")
			return
		}
	}
	condition := player.Username + player.Password
	result := User{}
	_, err = con.Get(condition)
	if err != nil {
		fmt.Println("this record is not in memcache")
		err = collection.Find(bson.M{"user_name": player.Username, "user_password": player.Password}).One(&result)
		if err != nil {
			fmt.Println("用户名或密码错误")
			//fmt.Println(err)
			return
		}
		str := fmt.Sprintf("%v", result)
		item1 := &memcache.Item{Key: condition, Value: []byte(str), Expiration: 0}
		seterr := con.Set(item1)
		fmt.Println("set in success")
		if seterr != nil {
			fmt.Printf("failed to set item: %s", seterr)
		}
	} else {
		fmt.Println("load from memcache success")
	}

	if err == nil {
		if len(server.players) > 0 {
			b, _ := json.Marshal(server.players)
			players = string(b)
			//fmt.Println(players)
		} else {
			fmt.Println("No player online")
			//fmt.Println(players)
		}
		server.players = append(server.players, player)
		//fmt.Println("result:", result.UserName, result.UserPassword)
		return
	}
	return

}

func onces() {
	session, err := mgo.Dial(Url)
	if err != nil {
		panic(err)
		//fmt.Println("err")
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)
	db := session.DB("game")
	collection := db.C("user")

	index := mgo.Index{
		Key:      []string{"user_name"},
		Unique:   true,
		DropDups: true,
	}
	err = collection.EnsureIndex(index)
	if err != nil {
		panic(err)
		//fmt.Println("err")
	}
}

//退出登录
func (server *CenterServer) Logout(params string) error {
	for i, v := range server.players {
		if v.Username == params {
			if len(server.players) == 1 {
				server.players = make([]*Player, 0)
			} else if i == len(server.players)-1 {
				server.players = server.players[:i]
			} else if i == 0 {
				server.players = server.players[1:]
			} else {
				server.players = append(server.players[:i], server.players[i+1:]...)
			}
			return nil
		}
	}
	return errors.New("Player not found.")
}

func (server *CenterServer) LogoutDis(number string) error {
	for i, v := range server.players {
		if v.Number == number {
			if len(server.players) == 1 {
				server.players = make([]*Player, 0)
			} else if i == len(server.players)-1 {
				server.players = server.players[:i]
			} else if i == 0 {
				server.players = server.players[1:]
			} else {
				server.players = append(server.players[:i], server.players[i+1:]...)
			}
			return nil
		}
	}
	return errors.New("Player not found.")
}

//准备
func (server *CenterServer) Ready(params string) (state string, err error) {
	var i int
	for _, v := range server.players {
		if v.Username == params {
			v.IsReady = 1
		}
		if v.IsReady == 1 {
			i = i + 1
		}
	}
	if i == 4 {
		state = "allready"
		for _, v := range server.players {
			v.IsReady = 0
		}
		return
	}
	return

}

func (server *CenterServer) Handle(method, params string) *ipc.Response {
	switch method {

	case "Login":
		players, err := server.Login(params)
		if err != nil {
			fmt.Println(err)
			return &ipc.Response{Code: err.Error()}
		}
		return &ipc.Response{"200", players}

	case "Logout":
		err := server.Logout(params)
		if err != nil {
			fmt.Println(err)
			return &ipc.Response{Code: err.Error()}
		}
		return &ipc.Response{Code: "200"}
	case "Ready":
		state, err := server.Ready(params)
		if err != nil {
			fmt.Println(err)
			return &ipc.Response{Code: err.Error()}
		}
		return &ipc.Response{"200", state}

	default:
		return &ipc.Response{Code: "404", Body: method + ":" + params}
	}
	return &ipc.Response{Code: "200"}
}
func (server *CenterServer) Name() string {
	return "CenterServer"
}
