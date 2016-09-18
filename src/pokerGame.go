package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/yuin/gopher-lua"
	"golang.org/x/crypto/scrypt"
	"ipc"
	"math/rand"
	"net"
	cg "poker/src/cg"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

var centerClient *cg.CenterClient
var secret = Room{}
var seeda, seedb, seed = 2, 2, 2 //逢人配
var seedname string
var rank string //每小局出完牌顺序排名
var timesa int
var timesb int
var hg string
var ml string

type Connection struct {
	State int
}

type Loginer struct {
	IsLogin int
}

type Users struct {
	User_Info []UserInfo
}

type UserInfo struct {
	Username string
	Number   string
	IsReady  int
}

type Room struct {
	Number1 bool
	Number2 bool
	Number3 bool
	Number4 bool
}

func StartCenterService() error {
	server := ipc.NewIpcServer(&cg.CenterServer{})
	client := ipc.NewIpcClient(server)
	centerClient = &cg.CenterClient{client}

	return nil
}

func TcpPipe(conn *net.TCPConn) {
	ipStr := conn.RemoteAddr().String()
	reader := bufio.NewReader(conn)

	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		handlers := GetCommandHandlers(conn)
		tokens := strings.Split(message, "+")
		fmt.Println(message)
		result := ""
		if handler, ok := handlers[tokens[0]]; ok {
			result = handler(tokens, conn)
			if result == "" {
				break
			}
			fmt.Println(result)
		} else {
			conn.Write([]byte("wrong request" + "\n"))
		}
		fmt.Println(conn.RemoteAddr().String() + ":" + string(message))
		if tokens[0] == "login" {
			if strings.Contains(result, "IsLogin\":1}") {
				loginname := strings.Split(result, "+")
				boradcastMessage(loginname[0]+"+"+loginname[2], conn)
			}
			result = result + "\n"
			conn.Write([]byte(result))
		} else if tokens[0] == "ready" {
			if result == "1" {
				boradcastMessage(message, conn)
				message = "allready" + "\n"
				b := []byte(message)
				for _, conns := range ConnMap {
					conns.Write(b)
				}
				Licensing(conn, seed, rank)
			} else {
				boradcastMessage(message, conn)
			}
		} else if tokens[0] == "logout" {
			boradcastMessage(message, conn)
		} else if tokens[0] == "play" {
			if result == "playGoOn" {
				boradcastMessageAll(message, conn)
			} else if result == "playOver" {
				message = "over" + "+" + rank + "\n"
				b := []byte(message)
				for _, conns := range ConnMap {
					conns.Write(b)
				}
			} else if result == "finish" {
				var winer string
				if string(rank[0]) == "1" || string(rank[0]) == "3" {
					winer = "[" + "1" + "," + "3" + "]"
				} else if string(rank[0]) == "2" || string(rank[0]) == "4" {
					winer = "[" + "2" + "," + "4" + "]"
				}
				rank = ""
				hg = ""
				seeda, seedb, seed = 2, 2, 2
				timesa, timesb = 0, 0
				seedname = ""
				message = "finish" + "+" + winer + "\n"
				b := []byte(message)
				for _, conns := range ConnMap {
					conns.Write(b)
				}
			}
		} else if tokens[0] == "huangong" {
			var turn string
			turn = string(hg[1])
			if result == "huan" {
				if len(hg) == 5 {
					if len(rank) == 6 {
						message = "huangongover" + "+" + turn + "\n"
						b := []byte(message)
						for _, conns := range ConnMap {
							conns.Write(b)
						}
						rank = ""
						hg = ""
					}
				} else if len(hg) == 3 {
					message = "huangongover" + "+" + turn + "\n"
					b := []byte(message)
					for _, conns := range ConnMap {
						conns.Write(b)
					}
					rank = ""
					hg = ""
				}
			}
		} else if tokens[0] == "sendmessage" {
			boradcastMessageAll(message, conn)
		}
		//最终执行的断开函数
		defer func() {
			rank = ""
			hg = ""
			seeda, seedb, seed = 2, 2, 2
			timesa, timesb = 0, 0
			seedname = ""
			ml = ""

			for k, _ := range ConnMap {
				if e := recover(); e != nil {
					fmt.Printf("Panicing: %s\r\n", e)
				}

				//fmt.Println(beginstr)
				if strings.Contains(k, conn.RemoteAddr().String()) {

					//fmt.Println(beginstr)
					if strings.Contains(result, "IsLogin\":1}") {
						loginon := strings.Split(result, "+")
						//boradcastMessage(message, conn)
						Disconnect(loginon[0], loginon[2], conn)
						message := "logout" + "+" + "{\"user\":{\"user_name\":\"" + loginon[0] + "\"}}" + "+" + loginon[2]
						boradcastMessage(message, conn)
						break
					} else if strings.Contains(result, "IsLogin\":0}") {
						fmt.Println("disconnected :" + ipStr)
						delete(ConnMap, k)
						conn.Close()
						break
					} else if result == "logout" {
						break
					} else if (result == "0") || (result == "1") {
						js, _ := simplejson.NewJson([]byte(tokens[1]))
						name := js.Get("user").Get("user_name").MustString()
						//boradcastMessage(message, conn)
						Disconnect(name, tokens[2], conn)
						message := "logout" + "+" + "{\"user\":{\"user_name\":\"" + name + "\"}}" + "+" + tokens[2]
						boradcastMessage(message, conn)
						break
					} else if result == "playGoOn" || result == "playOver" || result == "finish" {
						name := strings.TrimSpace(tokens[6])
						Disconnect(name, tokens[4], conn)
						message := "logout" + "+" + "{\"user\":{\"user_name\":\"" + name + "\"}}" + "+" + tokens[4]
						boradcastMessage(message, conn)
						break
					} else if result == "huan" {
						Disconnectnum(tokens[2], conn)
						break
					} else if result == "sendMessage" {
						nameContext := strings.TrimSpace(tokens[1])
						nameContexts := strings.Split(nameContext, ":")
						name := strings.TrimSpace(nameContexts[0])
						number := strings.TrimSpace(tokens[2])
						Disconnect(name, number, conn)
					}
				}
			}
		}()
	}
}

func GetCommandHandlers(conn *net.TCPConn) map[string]func(args []string, conn *net.TCPConn) string {
	return map[string]func([]string, *net.TCPConn) string{
		"login":       Login,
		"ready":       Ready,
		"play":        Play,
		"logout":      Logout,
		"huangong":    HuanGong,
		"sendmessage": SendMessage,
	}
}

func boradcastMessage(message string, conn *net.TCPConn) {
	message = message + "\n"
	b := []byte(message)

	// 遍历所有客户端并发送消息
	for _, conns := range ConnMap {
		if conns != conn {
			conns.Write(b)
		}
	}
}
func boradcastMessageAll(message string, conn *net.TCPConn) {
	message = message + "\n"
	b := []byte(message)
	// 遍历所有客户端并发送消息
	for _, conns := range ConnMap {
		conns.Write(b)
	}
}
func Login(args []string, conn *net.TCPConn) string {
	if len(ConnMap) < 4 {
		fmt.Println("A client connected : " + conn.RemoteAddr().String())
		ml = ml + "I"
		beginstr := ml + conn.RemoteAddr().String()
		// 新连接加入map
		ConnMap[beginstr] = conn
		//处理json
		js, err := simplejson.NewJson([]byte(args[1]))
		if err != nil {
			fmt.Println("json format error")
		}
		name := js.Get("user").Get("user_name").MustString()
		passwd := js.Get("user").Get("user_password").MustString()
		//密码加密
		passwdm := scrypts(passwd, "@#$%")
		var id = 0
		value := reflect.ValueOf(&secret).Elem()
		for i := 0; i < value.NumField(); i++ {
			if !(value.Field(i).Interface()).(bool) {
				id = i + 1
				break
			}
		}
		ids := strconv.Itoa(id)
		player := cg.NewPlayer()
		player.Username = name
		player.Password = passwdm
		player.IsReady = 0
		player.Number = ids
		//调用
		ps, err := centerClient.Login(player)
		var userInfos Users
		if err != nil { //登录失败，返回登录状态：0,和位置号
			fmt.Println("Failed login", err)
			login := Loginer{IsLogin: 0}
			loginjs, _ := json.Marshal(login)
			return string(loginjs) + "+" + "0"
			//conn.Write([]byte(string(loginjs) + "+" + string(ids)))
		} else { //登录成功，返回登录状态：1，和位置号
			value := reflect.ValueOf(&secret).Elem()
			//types := reflect.TypeOf(&secret).Elem()
			for i := 0; i < value.NumField(); i++ {
				if !(value.Field(i).Interface()).(bool) {
					value.Field(i).SetBool(true)
					break
				}
			}

			for _, v := range ps {
				userInfo := UserInfo{Username: v.Username, Number: v.Number, IsReady: v.IsReady}
				userInfos.User_Info = append(userInfos.User_Info, userInfo)
			}
			userInfosjs, _ := json.Marshal(userInfos.User_Info)
			login := Loginer{IsLogin: 1}
			loginjs, _ := json.Marshal(login)
			return string(name) + "+" + string(loginjs) + "+" + string(ids) + "+" + string(userInfosjs)
		}

	} else { //连接失败，返回连接状态：0
		connected := Connection{State: 0}
		connectedjs, _ := json.Marshal(connected)
		return string(connectedjs) + "+" + "0"
		//conn.Write([]byte(string(connectedjs) + "+" + "0"))
		fmt.Println("Already exist four players, join failed!")
	}
	return ""

}

//密码加密函数
func scrypts(str, salt string) string {
	kk, _ := scrypt.Key([]byte(str), []byte(salt), 16384, 8, 1, 32)
	return fmt.Sprintf("%x", kk)
}

//准备
func Ready(args []string, conn *net.TCPConn) string {
	js, err := simplejson.NewJson([]byte(args[1]))
	if err != nil {
		fmt.Println("json format error")
	}
	name := js.Get("user").Get("user_name").MustString()

	state, err := centerClient.Ready(name)
	if err != nil {
		return "err"
	}
	if state == "allready" {
		return "1"
	}
	return "0"
}

//发牌
func Licensing(conn *net.TCPConn, seed int, rank string) string {
	Carders := make(map[int]string)
	L := lua.NewState()
	defer L.Close()
	if err := L.DoFile("lua/PokerClass.lua"); err != nil {
		panic(err)
	}
	if err := L.CallByParam(lua.P{
		Fn:      L.GetGlobal("ctor"),
		NRet:    1,
		Protect: true,
	}, lua.LNumber(seed), lua.LString(rank)); err != nil {
		panic(err)
	}

	lvalue1 := L.GetGlobal("handstr1")
	lvalue2 := L.GetGlobal("handstr2")
	lvalue3 := L.GetGlobal("handstr3")
	lvalue4 := L.GetGlobal("handstr4")
	hgs := L.GetGlobal("hgs")
	hg = fmt.Sprintf("%s", hgs)

	Carders[0] = fmt.Sprintf("%s", lvalue1)
	Carders[1] = fmt.Sprintf("%s", lvalue2)
	Carders[2] = fmt.Sprintf("%s", lvalue3)
	Carders[3] = fmt.Sprintf("%s", lvalue4)

	k1 := strings.Split(Carders[0], " ")
	k2 := strings.Split(Carders[1], " ")
	k3 := strings.Split(Carders[2], " ")
	k4 := strings.Split(Carders[3], " ")
	klen1 := strconv.Itoa(len(k1) - 1)
	klen2 := strconv.Itoa(len(k2) - 1)
	klen3 := strconv.Itoa(len(k3) - 1)
	klen4 := strconv.Itoa(len(k4) - 1)

	nums := "[" + klen1 + "," + klen2 + "," + klen3 + "," + klen4 + "]"

	if seeda == 2 && seedb == 2 && seed == 2 {
		// 遍历所有客户端并发送消息
		var i int
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		turn := r.Intn(4) + 1
		turns := strconv.Itoa(turn)
		if turns == "1" || turns == "3" {
			seed = seeda
			seedname = "seeda"
		} else if turns == "2" || turns == "4" {
			seed = seedb
			seedname = "seedb"
		}

		for _, conns := range ConnMap {
			err := SendCard(Carders, conns, i, turns, nums)
			if err != nil {
				return "err"
			}
			i = i + 1
		}
		return "1"
	} else {

		var keys []string
		for k := range ConnMap {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var i int
		for _, conns := range keys {
			err := HuanGongCard(Carders, ConnMap[conns], hg, i, nums)
			if err != nil {
				return "err"
			}
			i = i + 1
		}
	}
	return ""
}
func SendMessage(args []string, conn *net.TCPConn) string {

	return "sendMessage"
}
func HuanGongCard(Carders map[int]string, conn *net.TCPConn, hg string, i int, nums string) error {
	var message string
	var turn string

	if len(hg) == 0 {
		turn = string(rank[0])
		rank = ""
	} else {
		turn = string(hg[1])
	}
	seedas := strconv.Itoa(seeda)
	seedbs := strconv.Itoa(seedb)
	seeds := strconv.Itoa(seed)
	specard := "[" + seedas + "," + seedbs + "," + seeds + "]"
	message = "huangong" + "+" + Carders[i] + "+" + turn + "+" + hg + "+" + specard + "+" + nums
	message = message + "\n"
	b := []byte(message)
	conn.Write(b)
	return nil

}

func SendCard(Carders map[int]string, conn *net.TCPConn, i int, turns, nums string) error {
	var message string
	specard := "[" + "2" + "," + "2" + "," + "2" + "]"
	message = "sendcard" + "+" + Carders[i] + "+" + turns + "+" + specard + "+" + nums
	message = message + "\n"
	b := []byte(message)
	conn.Write(b)
	return nil
}

func HuanGong(args []string, conn *net.TCPConn) string {
	huancard := args[1]
	huanseat := strings.TrimSpace(args[2])
	var receseat string

	if len(hg) == 5 {
		if huanseat == string(rank[0]) {
			receseat = string(hg[1])

			rank = rank + "1"
		} else if huanseat == string(rank[1]) {
			receseat = string(hg[3])
			rank = rank + "2"
		}
	} else if len(hg) == 3 {
		receseat = string(hg[1])
	}
	message := "rececard" + "+" + huancard + "+" + receseat
	message = message + "\n"
	b := []byte(message)
	fmt.Println(message)
	for _, conns := range ConnMap {
		conns.Write(b)
	}
	return "huan"
}

var Lasthandnum = make([]int, 5)

//Lasthandnum := make([]int,4)
//打牌过程处理函数
func Play(args []string, conn *net.TCPConn) string {
	seat := args[4]
	handnum := args[5]
	seats, _ := strconv.Atoi(seat)
	handnums, _ := strconv.Atoi(handnum)
	Lasthandnum[seats] = handnums
	if handnum == "0" {
		rank = rank + seat
		fmt.Println(rank)
		if len(rank) == 2 {
			if rank == "13" || rank == "31" {
				if Lasthandnum[2] > Lasthandnum[4] {
					rank = rank + "4"
				} else {
					rank = rank + "2"
				}
			} else if rank == "24" || rank == "42" {
				if Lasthandnum[1] > Lasthandnum[3] {
					rank = rank + "3"
				} else {
					rank = rank + "1"
				}
			}
		}
		if len(rank) == 3 {
			fmt.Println(rank)
			if rank == "123" || rank == "143" || rank == "321" || rank == "341" {
				if rank == "123" {
					rank = rank + "4"
				} else if rank == "143" {
					rank = rank + "2"
				} else if rank == "321" {
					rank = rank + "4"
				} else if rank == "341" {
					rank = rank + "2"
				}
				if seeda == 1 && seedname == "seeda" && seedb != 1 {
					if timesa < 3 {
						return "finish"
					}
				} else if seeda == 1 && seedname == "seedb" && seedb != 1 {
					seed = seeda
					seedname = "seeda"
				} else if seedb == 1 && seedname == "seedb" && seeda != 1 {
					timesb = timesb + 1
					if timesb >= 3 {
						seedb = 2
						timesb = 0
					}
					seeda = seeda + 2
					seed = seeda
					seedname = "seeda"
				} else if seedb == 1 && seedname == "seeda" && seeda != 1 {
					seeda = seeda + 2
					seed = seeda
					seedname = "seeda"
				} else if seedb == 1 && seeda == 1 && seedname == "seeda" {
					if timesa < 3 {
						return "finish"
					}
				} else if seedb == 1 && seeda == 1 && seedname == "seedb" {
					timesb = timesb + 1
					if timesb >= 3 {
						seedb = 2
						timesb = 0
					}
					seed = seeda
					seedname = "seeda"
				} else {
					seeda = seeda + 2
					seed = seeda
					seedname = "seeda"
				}
			} else if rank == "124" || rank == "142" || rank == "324" || rank == "342" {
				if rank == "124" {
					rank = rank + "3"
				} else if rank == "142" {
					rank = rank + "3"
				} else if rank == "324" {
					rank = rank + "1"
				} else if rank == "342" {
					rank = rank + "1"
				}

				if seeda == 1 && seedname == "seeda" && seedb != 1 {
					timesa = timesa + 1
					if timesa >= 3 {
						seeda = 2
						timesa = 0
					}
					seed = seeda
					seedname = "seeda"
				} else if seeda == 1 && seedname == "seedb" && seedb != 1 {
					seed = seeda
					seedname = "seeda"
				} else if seedb == 1 && seedname == "seedb" && seeda != 1 {
					timesb = timesb + 1
					if timesb >= 3 {
						seedb = 2
						timesb = 0
					}
					seeda = seeda + 1
					seed = seeda
					seedname = "seeda"
				} else if seedb == 1 && seedname == "seeda" && seeda != 1 {
					seeda = seeda + 1
					seed = seeda
					seedname = "seeda"
				} else if seedb == 1 && seeda == 1 && seedname == "seeda" {
					timesa = timesa + 1
					if timesa >= 3 {
						seeda = 2
						timesa = 0
					}
					seed = seeda
					seedname = "seeda"
				} else if seedb == 1 && seeda == 1 && seedname == "seedb" {
					timesb = timesb + 1
					if timesb >= 3 {
						seedb = 2
						timesb = 0
					}
					seed = seeda
					seedname = "seeda"
				} else {
					seeda = seeda + 1
					seed = seeda
					seedname = "seeda"
				}
			} else if rank == "132" || rank == "134" || rank == "312" || rank == "314" {
				if rank == "132" {
					rank = rank + "4"
				} else if rank == "134" {
					rank = rank + "2"
				} else if rank == "312" {
					rank = rank + "4"
				} else if rank == "314" {
					rank = rank + "2"
				}

				if seeda == 1 && seedname == "seeda" && seedb != 1 {
					if timesa < 3 {
						return "finish"
					}
				} else if seeda == 1 && seedname == "seedb" && seedb != 1 {
					seed = seeda
					seedname = "seeda"
				} else if seedb == 1 && seedname == "seedb" && seeda != 1 {
					timesb = timesb + 1
					if timesb >= 3 {
						seedb = 2
						timesb = 0
					}
					seeda = seeda + 3
					seed = seeda
					seedname = "seeda"
				} else if seedb == 1 && seedname == "seeda" && seeda != 1 {
					seeda = seeda + 3
					seed = seeda
					seedname = "seeda"
				} else if seedb == 1 && seeda == 1 && seedname == "seeda" {
					if timesa < 3 {
						return "finish"
					}
				} else if seedb == 1 && seeda == 1 && seedname == "seedb" {
					timesb = timesb + 1
					if timesb >= 3 {
						seedb = 2
						timesb = 0
					}
					seed = seeda
					seedname = "seeda"
				} else {
					seeda = seeda + 3
					seed = seeda
					seedname = "seeda"
				}
			} else if rank == "213" || rank == "231" || rank == "413" || rank == "431" {
				if rank == "213" {
					rank = rank + "4"
				} else if rank == "231" {
					rank = rank + "4"
				} else if rank == "413" {
					rank = rank + "2"
				} else if rank == "431" {
					rank = rank + "2"
				}

				if seeda == 1 && seedname == "seeda" && seedb != 1 {
					timesa = timesa + 1
					if timesa >= 3 {
						seeda = 2
						timesa = 0
					}
					seedb = seedb + 1
					seed = seedb
					seedname = "seedb"
				} else if seeda == 1 && seedname == "seedb" && seedb != 1 {
					seedb = seedb + 1
					seed = seedb
					seedname = "seedb"
				} else if seedb == 1 && seedname == "seedb" && seeda != 1 {
					timesb = timesb + 1
					if timesb >= 3 {
						seedb = 2
						timesb = 0
					}
					seed = seedb
					seedname = "seedb"
				} else if seedb == 1 && seedname == "seeda" && seeda != 1 {
					seed = seedb
					seedname = "seedb"
				} else if seedb == 1 && seeda == 1 && seedname == "seeda" {
					timesa = timesa + 1
					if timesa >= 3 {
						seeda = 2
						timesa = 0
					}
					seed = seedb
					seedname = "seedb"
				} else if seedb == 1 && seeda == 1 && seedname == "seedb" {
					timesb = timesb + 1
					if timesb >= 3 {
						seedb = 2
						timesb = 0
					}
					seed = seedb
					seedname = "seedb"
				} else {
					seedb = seedb + 1
					seed = seedb
					seedname = "seedb"
				}
			} else if rank == "214" || rank == "234" || rank == "412" || rank == "432" {
				if rank == "214" {
					rank = rank + "3"
				} else if rank == "234" {
					rank = rank + "1"
				} else if rank == "412" {
					rank = rank + "3"
				} else if rank == "432" {
					rank = rank + "1"
				}
				if seeda == 1 && seedname == "seeda" && seedb != 1 {
					timesa = timesa + 1
					if timesa >= 3 {
						seeda = 2
						timesa = 0
					}
					seedb = seedb + 2
					seed = seedb
					seedname = "seedb"
				} else if seeda == 1 && seedname == "seedb" && seedb != 1 {
					seedb = seedb + 2
					seed = seedb
					seedname = "seedb"
				} else if seedb == 1 && seedname == "seedb" && seeda != 1 {
					if timesb < 3 {
						return "finish"
					}
				} else if seedb == 1 && seedname == "seeda" && seeda != 1 {
					seed = seedb
					seedname = "seedb"
				} else if seedb == 1 && seeda == 1 && seedname == "seeda" {
					timesa = timesa + 1
					if timesa >= 3 {
						seeda = 2
						timesa = 0
					}
					seed = seedb
					seedname = "seedb"
				} else if seedb == 1 && seeda == 1 && seedname == "seedb" {
					if timesb < 3 {
						return "finish"
					}
				} else {
					seedb = seedb + 2
					seed = seedb
					seedname = "seedb"
				}
			} else if rank == "241" || rank == "243" || rank == "421" || rank == "423" {
				if rank == "241" {
					rank = rank + "3"
				} else if rank == "243" {
					rank = rank + "1"
				} else if rank == "421" {
					rank = rank + "3"
				} else if rank == "423" {
					rank = rank + "1"
				}
				if seeda == 1 && seedname == "seeda" && seedb != 1 {
					timesa = timesa + 1
					if timesa >= 3 {
						seeda = 2
						timesa = 0
					}
					seedb = seedb + 3
					seed = seedb
					seedname = "seedb"
				} else if seeda == 1 && seedname == "seedb" && seedb != 1 {
					seedb = seedb + 3
					seed = seedb
					seedname = "seedb"
				} else if seedb == 1 && seedname == "seedb" && seeda != 1 {
					if timesb < 3 {
						return "finish"
					}
				} else if seedb == 1 && seedname == "seeda" && seeda != 1 {
					seed = seedb
					seedname = "seedb"
				} else if seedb == 1 && seeda == 1 && seedname == "seeda" {
					timesa = timesa + 1
					if timesa >= 3 {
						seeda = 2
						timesa = 0
					}
					seed = seedb
					seedname = "seedb"
				} else if seedb == 1 && seeda == 1 && seedname == "seedb" {
					if timesb < 3 {
						return "finish"
					}
				} else {
					seedb = seedb + 3
					seed = seedb
					seedname = "seedb"
				}
			}
			if seeda > 13 {
				seeda = 1
				seed = seeda
				seedname = "seeda"
			}
			if seedb > 13 {
				seedb = 1
				seed = seedb
				seedname = "seedb"
			}
			return "playOver"
		}
		return "playGoOn"
	} else {
		return "playGoOn"
	}
	return "playGoOn"
}

//退出登录函数
func Logout(args []string, conn *net.TCPConn) string {

	js, err := simplejson.NewJson([]byte(args[1]))
	if err != nil {
		fmt.Println("json format error")
	}
	name := js.Get("user").Get("user_name").MustString()

	err = centerClient.Logout(name)
	if err != nil {
		return "err"
	}
	num := strings.TrimSpace(args[2])
	id, _ := strconv.Atoi(num)
	id = id - 1
	value := reflect.ValueOf(&secret).Elem()
	for i := 0; i < value.NumField(); i++ {
		if i == id {
			value.Field(i).SetBool(false)
			break
		}
		//fmt.Printf("Field %v: %v\n", i, value.Field(i))
	}
	ipStr := conn.RemoteAddr().String()
	fmt.Println("logout:" + ipStr)
	for k, _ := range ConnMap {
		if strings.Contains(k, conn.RemoteAddr().String()) {
			delete(ConnMap, k)
			break
		}
	}
	conn.Close()

	return "logout"
}

//恶性关闭时的断开函数
func Disconnect(username, number string, conn *net.TCPConn) {
	fmt.Println("=========================Disconnect")
	_ = centerClient.Logout(username)
	num := strings.TrimSpace(number)
	id, _ := strconv.Atoi(num)
	id = id - 1
	value := reflect.ValueOf(&secret).Elem()
	for i := 0; i < value.NumField(); i++ {
		if i == id {
			value.Field(i).SetBool(false)
			break
		}
	}
	ipStr := conn.RemoteAddr().String()
	fmt.Println("disconnected:" + ipStr)
	//fmt.Println(beginstr)
	for k, _ := range ConnMap {
		if strings.Contains(k, conn.RemoteAddr().String()) {
			delete(ConnMap, k)

			break
		}
	}
	conn.Close()
}

func Disconnectnum(number string, conn *net.TCPConn) {
	fmt.Println("=========================Disconnectnum")
	_ = centerClient.LogoutDis(number)
	num := strings.TrimSpace(number)
	id, _ := strconv.Atoi(num)
	id = id - 1
	value := reflect.ValueOf(&secret).Elem()
	for i := 0; i < value.NumField(); i++ {
		if i == id {
			value.Field(i).SetBool(false)
			break
		}
	}
	ipStr := conn.RemoteAddr().String()
	fmt.Println("disconnected:" + ipStr)
	//fmt.Println(beginstr)
	for k, _ := range ConnMap {
		if strings.Contains(k, conn.RemoteAddr().String()) {
			delete(ConnMap, k)
			break
		}
	}
	conn.Close()
}
