/* 
	wrapper elements to handle this application's specific redis clients' needs
*/
package main

import (
	"time"
	"github.com/go-redis/redis"
)

type redisclient struct{
	client *redis.Client
}

func NewPktCacheClient()(redisclient){

	return redisclient{
		client: redis.NewClient(&redis.Options{
                                Addr: "localhost:6379",
                                DB: 0,
                        },
	}
}

func (r redisclient)ReConnect()(bool){

	for i:=0;i<5;i++{
		// close the client
		r.client.Close()

		// give it some time before retrying to conenct back into the cache
		time.Sleep(1*time.Second)

		// create a new connection to the cache
		r.client := NewPktCacheClient()

		// verify that the client was properly setup/connected
		if r.client.IsConnected()!=false{
			return true
		}
	}
	r.client.Close()
	return false
}

func (r redisclient)IsConnected()(bool){
	if _,err:=r.client.Ping().Result(); err!=nil{
		// client is no longer connected !
		return false
	}else{
		return true
	}
}

func (r redisclient)Close(){
	r.client.Close()
}
