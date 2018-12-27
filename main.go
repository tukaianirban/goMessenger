package main

import(
	"fmt"
	"net"
)

func main(){

	// cache of client identities as map of client-identity and corresponding worker struct
	clients:= NewClientsCache()

	// afford max 10 accepted but unhandled connections
	chconn := make(chan net.Conn, 10)
	chserverclose := make(chan bool)	// unbuffered channel

	// start the server
	go StartServer("tcp4", ":8881", chconn, chserverclose)
	fmt.Println("Server started at local machine")

	// create ingress and egress packet pools; start the postman goroutine
	ingresspktpool := make(chan chan Packet, 50)
	chdispatcherquit := make(chan bool)
	go postman(ingresspktpool, clients, chdispatcherquit)

	// for every new client connection, assign a read/write channel and a worker goroutine
	for{
		clientchannel:= NewChannel(<-chconn)
		worker:=NewWorker(clientchannel, ingresspktpool)
		fmt.Println("New worker created to serve new client connection")
		// cache the channel and its identity and init the channel
		if i,err:=worker.IdentifyClient(); err!=nil{
			fmt.Println("Unable to recognize client at ", clientchannel.conn.RemoteAddr(), " Closing down the connection")
			worker.Stop()
		}else{
			fmt.Println("New client identified as=", i)
			clients.Add(i, &worker)
			go worker.Start()
		}
	}

	// block on a dummy channel (leaky way !!!!!)
	chdummy:=make(chan bool)
	<-chdummy

	// time to end the application

	// clean up the cache of client identities
	fmt.Println("Time to end the application")
	for _,key := range clients.GetKeys(){
		clients.Delete(key)
	}

	// close downt the ingress packet pool
	close(ingresspktpool)
	close(chdispatcherquit)
}
