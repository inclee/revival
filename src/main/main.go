package main

import (
	"com/inclee/revival/net"
	"fmt"
	"time"
)

type Hello struct{
	Name string
}

func (h Hello)String() string {
	return "Hi Hi ,U Recviced me " + h.Name
}
type newSession func(session *net.Session)

func (newSession)OnLinked(session *net.Session)  {
	onSession(session)
}
func onSession(session *net.Session)  {
	for {
		msg,err := session.Receive()
		if err != nil {
			fmt.Println("on session recv error " + err.Error())
		}else{
			fmt.Println(msg.(*Hello))
		}
	}
}


func main() {
	json := net.Json()
	json.Register(Hello{})
	server := net.NewServer("tcp",json,"0.0.0.0:9888",newSession(onSession))
	go func() {
		server.Start()
	}()

	conn , err := net.Dial("tcp","127.0.0.1:9888",json)
	if err != nil {
		fmt.Printf("client dial error %s \n",err)
	}
	fmt.Println("connect to server ")
	for i :=1;i <=10 ;i++ {
		h := Hello{Name:fmt.Sprintf("inclee %d",i)}
		fmt.Println("send to server: " + h.Name)
		conn.Send(h)
		time.Sleep(1* time.Second)
	}
}
