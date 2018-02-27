package net

import (
	"net"
	"log"
	"encoding/json"
	"fmt"
	"reflect"
)

type Server struct {
	network string
 	protocol Protocol
 	addr  string
 	sessions []*Session
 	linkHandler LinkHandler
}

type Coder interface{
	Receive()(msg interface{},err error)
	Send(msg interface{}) (err error)
}
type Protocol interface{
	NewCoder(conn net.Conn)(Coder)
}

type LinkHandler interface{
	OnLinked(session *Session)
}

func NewServer(network string,protocol Protocol,addr string,handler LinkHandler)Server {
	return Server{network:network,protocol:protocol,addr:addr,linkHandler:handler}
}

func (s Server)Start()  {
	listener,err := net.Listen(s.network,s.addr)
	fmt.Println("listening at " + s.addr)
	if err != nil{
		log.Println(err)
	}else {
		for{
			conn,err :=listener.Accept()
			fmt.Println("accpet new conn")
			if err != nil {
				log.Println(err)
			}
			go func() {
				session := NewSession(conn,s.protocol)
				s.sessions = append(s.sessions,session)
				s.linkHandler.OnLinked(session)
				log.Printf("sessions count %d",len(s.sessions))
			}()
		}
	}
}

func NewSession(conn net.Conn,protocol Protocol) (*Session) {
	return &Session{conn:conn,coder: protocol.NewCoder(conn)}
}

func Dial(network string, address string,protocol Protocol) (*Session,error){
	conn ,err  := net.Dial(network,address)
	if err != nil {
		return nil, err
	}
	return NewSession(conn,protocol) ,nil
}

type Session struct {
	conn net.Conn
	coder Coder
}

type jsonIn struct {
	Header string
	Body json.RawMessage
}

type jsonOut struct {
	Header string
	Body interface{}
}

func (s Session)Receive()(msg interface{},err error){
	return s.coder.Receive()
}

func (s Session)Send(msg interface{})  {
	s.coder.Send(msg)
}

type JsonProtocol struct{
	types map[string] reflect.Type
	names map[reflect.Type]string
}

func Json()* JsonProtocol  {
	return &JsonProtocol{
		types:make(map[string]reflect.Type),
		names:make(map[reflect.Type]string),
	}
}
func (j *JsonProtocol)Register(msg interface{})  {
	t := reflect.TypeOf(msg)
	name := t.PkgPath() + "/" + t.Name()
	j.types[name] = t
	j.names[t] = name
}

func (j *JsonProtocol)NewCoder(conn net.Conn ) (coder Coder) {
	return &JsonCoder{j,json.NewDecoder(conn),json.NewEncoder(conn)}
}

type JsonCoder struct {
	p * JsonProtocol
	decoder *json.Decoder
	encoder *json.Encoder
}
func (c *JsonCoder)Receive()(msg interface{},err error){
	var in jsonIn
	err = c.decoder.Decode(&in)
	if err != nil {
		return nil,err
	}
	var body interface{}
	if in.Header != ""{
		if t,exists := c.p.types[in.Header]; exists{
			body = reflect.New(t).Interface()
		}
	}
	err = json.Unmarshal(in.Body,&body)
	if err != nil {
		return nil,err
	}
	return body,nil
}
func (c *JsonCoder)Send(msg interface{}) (err error){
	var out jsonOut
	t := reflect.TypeOf(msg)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if name, exists := c.p.names[t]; exists {
		out.Header = name
	}
	out.Body = msg
	return c.encoder.Encode(&out)
}


