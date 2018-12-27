package main

/* a worker takes ingress packets from the channel, decodes it and dispatches to the ingress message queue in postman
  it takes egress packets, encodes them and sends them to its egress channel

*/

import(
	"fmt"
)

type Worker struct{
	identity string				// remote client identity
	networkchan *Channel
	ingresspktpool chan chan Packet
	chpktingress chan Packet
	chpktegress chan Packet
	chquit chan bool
	isActive bool
}

func NewWorker(ch *Channel, inpktpool chan chan Packet)(Worker){

	ch.Init()
	return Worker{
		identity: "",
		networkchan: ch,
		ingresspktpool: inpktpool,
		chpktingress: make(chan Packet, 5),
		chpktegress: make(chan Packet, 5),
		chquit: make(chan bool),
		isActive:true,
	}
}

func (w *Worker)Start(){
	defer w.Stop()

	flagStop:=false

	for !flagStop{
		select{
			// ingress pkt from network
			case pktbytes,ok:= <-w.networkchan.ichannel:
				if !ok{
//					fmt.Println("Worker detected network ingress channel is closed. Closing down the worker!")
					flagStop=true
					break
				}
				var pkt Packet
				if err:=DecodePacket(pktbytes, &pkt); err!=nil{
					fmt.Println("Error reading Packet from client(", w.networkchan.conn.RemoteAddr().String(),"). Reason=", err.Error())
					break
				}
//				fmt.Println("Worker got new packet from client=", pkt)
				w.chpktingress<- pkt
				w.ingresspktpool<- w.chpktingress


			// egress pkt recvd from dispatcher
			case pkt,ok:= <-w.chpktegress:
				if !ok{
//					fmt.Println("Worker detected egress packet channel is closed. Closing down the worker!")
					flagStop=true
					break
				}
//				fmt.Println("Worker egress channel received packet to send=", pkt)
				pktbytes, err:=pkt.Encode()
				if err!=nil{
					fmt.Println("Error transforming egress packet (Worker=", w.identity, "). Reason=", err.Error())
					break
				}
//				fmt.Println("Worker egress packet after encoding=", pktbytes)
				w.networkchan.echannel<- pktbytes


			// quit signal received
			case fl:=<-w.chquit:
				fmt.Println("Worker:", w.identity," Received QUIT signal=", fl)
				if fl{
					return
				}
		}
	}
}

func (w *Worker)Stop(){

//	fmt.Println("Worker stop called")
//	fmt.Printf("Underlying channel at this stage=%+v\n", w.networkchan)
	if w.isActive{
		w.isActive=false

		// close the conn Channel
		w.networkchan.Close()

		// close the Packet channels towards the worker pools
		close(w.chpktingress)
		close(w.chpktegress)

		// close down the quit channel
		close(w.chquit)

//		fmt.Println("Worker closed")
	}

}


func (w *Worker)IdentifyClient()(string, error){

//	fmt.Println("Waiting to identify client")
        pktbytes,ok:= <-w.networkchan.ichannel
	if !ok{
//		fmt.Println("Worker detected network channel closed while waiting to identify client")
		w.Stop()
		return "", fmt.Errorf("Worker detected network channel closed while waiting to identify client")
	}
//	fmt.Println("Received identity packet bytes from client=", pktbytes)
        var pkt Packet
	if err:=DecodePacket(pktbytes, &pkt); err!=nil{
                fmt.Println("Error reading in identity pkt from client=", w.networkchan.conn.RemoteAddr().String(), " Reason=", err.Error())
                return "", err
        }
//	fmt.Printf("Identity packet from client=%+v\n", pkt)
	w.identity = string(pkt.srcIdentity[:])

        return w.identity, nil
}
