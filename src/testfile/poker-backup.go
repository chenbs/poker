package main

import (
	//"bufio"

	//"encoding/json"
	"fmt"
	//"github.com/bitly/go-simplejson"
	//"github.com/yuin/gopher-lua"
	//"golang.org/x/crypto/scrypt"
	//"ipc"
	//"math/rand"
	"net"
	//"poker/src/gy"
	//"reflect"
	//"strconv"
	//"strings"
	//"time"
	//"pokerGame"
)

// 用来记录所有的客户端连接
var ConnMap map[string]*net.TCPConn

/*var centerClient *cg.CenterClient
var secret = Room{}
var seeda, seedb, seed int //逢人配
var rank map[int]int       //每小局出完牌顺序排名
*/
/*type Connection struct {
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
}*/

/*type Readyer struct {
	IsReady int
}*/
/*type Room struct {
	Number1 bool
	Number2 bool
	Number3 bool
	Number4 bool
}*/

/*func startCenterService() error {
	server := ipc.NewIpcServer(&cg.CenterServer{})
	client := ipc.NewIpcClient(server)
	centerClient = &cg.CenterClient{client}

	return nil
}*/
func main() {

	fmt.Println("Starting the server ...")
	fmt.Println("-----version:1.0------")
	StartCenterService()
	var tcpAddr *net.TCPAddr
	ConnMap = make(map[string]*net.TCPConn)
	tcpAddr, _ = net.ResolveTCPAddr("tcp", "192.168.100.58:9999")
	tcpListener, err := net.ListenTCP("tcp", tcpAddr)
	defer tcpListener.Close()

	if err != nil {
		fmt.Println("Server starting failed...")
		panic(err)
	} else {
		fmt.Println("server is OK ,ready to connect")
	}

	for {
		tcpConn, err := tcpListener.AcceptTCP()
		if err != nil {
			continue
		}
		fmt.Println("A client connected... : " + tcpConn.RemoteAddr().String())
		go TcpPipe(tcpConn)
	}
}

/*func tcpPipe(conn *net.TCPConn) {

	ipStr := conn.RemoteAddr().String()

	reader := bufio.NewReader(conn)

	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			return
		}

		handlers := GetCommandHandlers(conn)
		tokens := strings.Split(message, "+")
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
				Licensing(conn, 2)
			} else {
				boradcastMessage(message, conn)
			}
		} else if tokens[0] == "logout" {
			boradcastMessage(message, conn)
		} else if tokens[0] == "play" {
			if result == "playGoOn" {
				boradcastMessage(message, conn)
			} else {

			}

		}
		//最终执行的断开函数
		defer func() {
			for k, _ := range ConnMap {
				if k == conn.RemoteAddr().String() {
					if strings.Contains(result, "IsLogin\":1}") {
						loginon := strings.Split(result, "+")
						boradcastMessage(message, conn)
						Disconnect(loginon[0], loginon[2], conn)
						break
					} else if strings.Contains(result, "IsLogin\":0}") {
						fmt.Println("disconnected :" + ipStr)
						delete(ConnMap, conn.RemoteAddr().String())
						conn.Close()
						break
					} else if result == "logout" {
						break
					} else if (result == "0") || (result == "1") {
						js, _ := simplejson.NewJson([]byte(tokens[1]))
						name := js.Get("user").Get("user_name").MustString()
						boradcastMessage(message, conn)
						Disconnect(name, tokens[2], conn)
						break
					}
				}
			}
		}()
	}
}

func GetCommandHandlers(conn *net.TCPConn) map[string]func(args []string, conn *net.TCPConn) string {
	return map[string]func([]string, *net.TCPConn) string{
		"login":  Login,
		"ready":  Ready,
		"play":   Play,
		"logout": Logout,
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

func Login(args []string, conn *net.TCPConn) string {
	if len(ConnMap) < 4 {
		fmt.Println("A client connected : " + conn.RemoteAddr().String())
		// 新连接加入map
		ConnMap[conn.RemoteAddr().String()] = conn
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
		//types := reflect.TypeOf(&secret).Elem()
		for i := 0; i < value.NumField(); i++ {
			if !(value.Field(i).Interface()).(bool) {
				id = i + 1
				break
			}
			//fmt.Printf("Field %v: %v\n", i, value.Field(i))
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
				//fmt.Printf("Field %v: %v\n", i, value.Field(i))
			}

			for _, v := range ps {
				userInfo := UserInfo{Username: v.Username, Number: v.Number, IsReady: v.IsReady}
				userInfos.User_Info = append(userInfos.User_Info, userInfo)
			}
			userInfosjs, _ := json.Marshal(userInfos.User_Info)
			login := Loginer{IsLogin: 1}
			loginjs, _ := json.Marshal(login)
			return string(name) + "+" + string(loginjs) + "+" + string(ids) + "+" + string(userInfosjs)
			//conn.Write([]byte(string(loginjs) + "+" + string(ids)))
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
func Licensing(conn *net.TCPConn, seed int) string {
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
	}); err != nil {
		panic(err)
	}

	if err := L.CallByParam(lua.P{
		Fn:      L.GetGlobal("paixu"),
		NRet:    1,
		Protect: true,
	}, lua.LNumber(seed)); err != nil {
		panic(err)
	}
	lv1 := L.GetGlobal("p1Value")
	lv2 := L.GetGlobal("p2Value")
	lv3 := L.GetGlobal("p3Value")
	lv4 := L.GetGlobal("p4Value")

	Carders = Dispose(lv1, Carders, 0)
	Carders = Dispose(lv2, Carders, 1)
	Carders = Dispose(lv3, Carders, 2)
	Carders = Dispose(lv4, Carders, 3)

	// 遍历所有客户端并发送消息
	var i int
	for _, conns := range ConnMap {
		err := SendCard(Carders, conns, i)
		if err != nil {
			return "err"
		}
		i = i + 1
	}

	return "1"
}

func SendCard(Carders map[int]string, conn *net.TCPConn, i int) error {
	var message string
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	turn := r.Intn(4) + 1
	message = "sendcard" + "+" + Carders[i] + "+" + turn + "+" + 2 + "+" + 2 + "+" + 2
	message = message + "\n"
	b := []byte(message)
	conn.Write(b)
	return nil
}

func Dispose(lv lua.LValue, Carders map[int]string, i int) map[int]string {
	card := fmt.Sprintf("%D", lv)
	var Cards string
	strcard := strings.Split(card, `) %`)
	for _, v := range strcard {
		vv := HandleCards(v)
		Cards = Cards + vv + " "
	}
	Carders[i] = Cards
	return Carders
}

//取出牌并返回
func HandleCards(params string) string {
	var vv string
	for _, v := range params {
		if v >= 48 && v <= 57 { //判断是否为0--9数字
			vv = vv + fmt.Sprintf("%c", v)
		}
	}
	return vv

}

//打牌过程处理函数
func Play(args []string, conn *net.TCPConn) string {

	chupai := args[1]
	chupaiPX := args[2]
	turn := args[3]
	seat := args[4]
	handnum := args[5]
	if handnum == 0 {

	} else {
		return "playGoOn"
	}

	return ""
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
	delete(ConnMap, conn.RemoteAddr().String())
	conn.Close()

	return "logout"
}

//恶性关闭时的断开函数
func Disconnect(username, number string, conn *net.TCPConn) {

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
	delete(ConnMap, conn.RemoteAddr().String())
	conn.Close()
}
*/
