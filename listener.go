package main

import(
	"net"
	"fmt"
)

func StartServer(inet string, addr string, chconn chan<- net.Conn, chdone chan bool){

	// create a listener
	listener, err:= net.Listen(inet, addr)
	if err!=nil{
		return
	}
	defer listener.Close()

	// start listening
	go serverlistener(chconn, listener)

	for{
		if fl:=<-chdone; fl{
			listener.Close()
			break
		}
	}

}

func serverlistener(chconn chan<- net.Conn, listener net.Listener){
	for {
		conn, err:=listener.Accept()
		if err!=nil{
			fmt.Println("Error accepting incoming connection. Reason=", err.Error())
		}else{
			chconn<- conn
			fmt.Println("New connection received from client=", conn.RemoteAddr().String())
		}
	}
	fmt.Println("Exiting server listener loop")
}
