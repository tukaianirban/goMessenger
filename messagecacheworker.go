/* message cache worker will interact with the (semi)persistent storage
It will store any user messages whose destination user is not connected to the system
It will on-request serve out user messages when a user's worker thread requests for it.

*/

package main

import (
	"fmt"
	"bytes"
)

/* message store getter worker is a type of worker that will receive 1 worker identity, 
   fetch all messages from cache that are destined to this worker,
   send the messages out to the chpktstouserworker channel,
   and then wait on for the next user identity
*/
func messagestoregetter(chuseridentity chan [4]byte, chpktstouserworker chan Packet){

	// connect to the cache
	// for now, redis cache is being used
	redisclient := redis.NewClient(&redis.Options{
				Addr: "localhost:6379",
				DB: 0,
			})
	defer redisclient.Close()

	var pktbuf bytes.Buffer
	gobdecoder := gob.NewDecoder(&pktbuf)
	for{
		// before picking up the next user identity to serve, make sure to check the cache connection
		if pong,err:= redisclient.Ping().Result(); err!=nil{
			fmt.Println("Error connecting to redisclient. Reason=", err.Error())
			break
		}

		// serve the next user identity
		userid := <-chuseridentity
		for{
			statuscmd:=redisclient.RPop(string(userid))
			if statuscmd.Err()!=nil{
				// irrespective of the type of error (for now)
				// return an indicator that no more packets are available in cache for this user identity
				fmt.Println("Error when fetching more packets for the user identity=", string(userid))
				fmt.Println("No more data fetches for this user!!!")

				//send signal to user worker indicating no more cached data available for this user
				chpktstouserworker<- Packet{}

				break
			}


			data,err := statuscmd.Bytes()
			if err!=nil{
				fmt.Println("Error converting cache-retrieved data to byte[]=", err.Error())
				continue
			}

			fmt.Println("User identity(", string(userid)," Next data=", data)
			if _,err:=pktbuf.Write(data); err!=nil{
				fmt.Println("Error temporarily writing-in the cache-retrieved data=", err.Error())
			}else{
				var pkt Packet
				if err:=gobdecoder.Decode(&pkt); err!=nil{
					fmt.Println("Error decoding the cache-retrieved data into a Packet format=", err.Error())
				}else{
					fmt.Println("Retrieved pkt for user-identity(", string(userid), "). Packet dump=", pkt)
					chpktstouserworker<- pkt
				}
				pktbuf.Reset()
			}
		}
	}// end of infinite loop


}


/* message store setter gets the next worker's data channel, extracts the Packet from it,
   encodes it to gob format and stores it into the cache
   It will retry storing a data value until and unless it succeeds

*/
func messagestoresetter(chuserpktchannels chan chan Packet){

	// connect to the cache
	// for now, redis cache is being used
	redisclient := NewPktCacheClient()
	defer redisclient.Close()

	// GOB encoder
	var pktbuf bytes.Buffer
	gobencoder := gob.NewEncoder(&pktbuf)

	var chdata chan Packet

	for{
		select{
			case chdata:= <-chuserpktchannels:

			case <-time.After(5*time.Second):
				// if for 5sec, no cache has been done, then re-check connectivity to the cache
				if _,err:=redisclient.Ping().Result(); err!=nil{
					fmt.Println("Cache setter worker detected connectivity loss of redis client!")

					// implement a retry mechanism !
					return
				}
		}
		if datapkt, flagchopen:= <-chdata; !flagchopen{
			fmt.Println("Encountered a closed worker data channel! Skipping it")
			continue
		}else{
			if err := gobencoder.Encode(datapkt); err!=nil{
				fmt.Println("Error encoding a user packet. Reason=", err.Error())
				fmt.Printf("Error Packet dump= %+v\n", datapkt)
				continue
			}
			intcmd:=redisclient.LPush(string(datapkt.dstIdentity),pktbuf.Bytes())
			if intcmd.Err()!=nil{
				fmt.Println("Error pushing in a data packet=", err.Error())
				fmt.Printf("Error packet dump=%+v\n", datapkt)

				// as a retry mechanism, put the packet back into the worker pkt channel to be extracted again later
				chdata<- datapkt
			}else{
				fmt.Println("Stored a new packet for dstIdentity=", string(datapkt.dstIdentity))
			}
		}
		pktbuf.Reset()
	}

}
