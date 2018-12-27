package main

import (
	"fmt"
)
// eats packets from ingress pool and feeds into egress pool
func postman(ingresspktpool chan chan Packet, cache *ClientsCache, chquit chan bool){

	for {
		select{
			// POSSIBLE POINT OF CRASH !!!!
			// IF USER DISCONNECTS SESSION (and therefore the Worker is closed) IMMEDIATELY AFTER UPLOADING A MSG !
			case pkt:= <-( <-ingresspktpool):
//				fmt.Println("Postman got new packet to dispatch. Dump=", pkt)
				w,err:=cache.GetWorker(string(pkt.dstIdentity[:]))
//				fmt.Println("Packet's destination=", string(pkt.dstIdentity[:]))
//				fmt.Println("Error in destination worker lookup=", err.Error())
//				fmt.Println("Packet's destination worker identity=", w.identity)
				if err!=nil{
					fmt.Println("Packet dropped : Could not find destination worker (identity: ", pkt.dstIdentity)
				}else{
//					fmt.Println("Postman dispatched packet to destination worker with identity=", w.identity)
					w.chpktegress<- pkt
				}
			case fl:= <-chquit:
//				fmt.Println("Postman received control signal")
				if fl{
					close(chquit)
					return
				}
		}
	}
}
