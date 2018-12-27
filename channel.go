package main

import (
	"io"
//	"io/ioutil"
	"bufio"
	"net"
	"fmt"
)

type Channel struct{
	conn net.Conn
	bufreader *bufio.Reader
	ichannel chan []byte
	echannel chan []byte
	isActive bool
}

func (c *Channel)reader(){
	for{

		// read each byte in, check if its the (custom)EOF or not
		pktheadersize:=16
		pktbuf:= make([]byte, pktheadersize)

		// read in the header sections
		if n,err:=c.bufreader.Read(pktbuf[:16]); err!=nil{
			fmt.Println("Channel reader: Error reading in the header of a new packet=", err.Error())
			if err==io.EOF{
				fmt.Println("Initiate channel closure")
				c.Close()
				return
			}else{
				fmt.Println("Channel reader: No action on this error")
			}
		}else{
			if n<16{
				fmt.Println("Channel reader: Corrupted or too small header read in. Discard packet")
				continue
			}
		}

//		pktbuf=append(pktbuf, tmpbuf[:16]...)
//		fmt.Println("New packet header read in=", pktbuf)
		// read in the packet payload
		if tmpbuf,err:=c.bufreader.ReadBytes(byte('\x00')); err!=nil{
			pktbuf = append(pktbuf, tmpbuf[:len(tmpbuf)-1]...)
			if err==io.EOF{
				c.ichannel<- pktbuf
				//close down the channel as the underlying net.Conn has encountered io.EOF
				c.Close()
				return
			}else{
				fmt.Println("Channel reader encountered (ignored)error when reading packet payload=", err.Error())
			}
		}else{
			// complete packet payload read in
			pktbuf = append(pktbuf, tmpbuf[:len(tmpbuf)-1]...)
			c.ichannel<- pktbuf
		}

/*
		for !flagfinished{
			var i int
			for i=0;i<fragbufsize;i++{
				if _,err:=c.conn.Read(tmpbuf[i:i+1]); err!=nil{
					fmt.Println("Error reading bytes from net.Conn=", err.Error())
//					fmt.Println("Initiate channel closure")
					c.Close()
					return
				}else{
					if tmpbuf[i]==byte('\x00'){
						// last fragment read in; packet received completely; send in packet without (custom)EOF
						if i>0{
							tmpbuf=tmpbuf[:i]
						}else{
							tmpbuf=tmpbuf[:0]
						}
						flagfinished=true
						break
					}
				}
			}
//			fmt.Println("Read in packet fragment=", tmpbuf, " size=", i)
			pktbuf = append(pktbuf, tmpbuf...)
		}
		c.ichannel<- pktbuf
*/
//		fmt.Println("Channel reader read in complete packet=", pktbuf)
	}
}

func (c *Channel)writer(){
	for{
		pktbytes,ok:= <-c.echannel
		if !ok{
			c.Close()
//			fmt.Println("Channel writer detected egress channel is closed. Closing down writer")
			return
		} 
//		fmt.Println("Channel writer sending out packet.Dump=", pktbytes)
		if _,err:=c.conn.Write(pktbytes); err!=nil{
			fmt.Println("Error when trying to send packet to the client. Reason=", err.Error())

			c.Close()
			return
			// !!!!!!!!!   MANY OTHER TYPES OF ERRORS CAN HAPPEN HERE. !!! NEEDS PROPER HANDLING
		}
	}
}

func (c *Channel)Close(){
	if c.isActive{
		c.isActive=false
		c.bufreader.Discard(c.bufreader.Buffered())
		c.conn.Close()
		close(c.ichannel)
		close(c.echannel)
//		fmt.Println("Channel closed")
	}
}

func (c *Channel)Init(){

	go c.reader()
	go c.writer()
}

func NewChannel(c net.Conn)(*Channel){
	return &Channel{
		conn:c,
		bufreader: bufio.NewReader(c),
		ichannel: make(chan []byte, 256),
		echannel: make(chan []byte, 256),
		isActive:true,
	}
}

