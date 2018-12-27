package main

import (
	"fmt"
	"sync"
)

type ClientsCache struct{
	lock *sync.RWMutex
	clients map[string]*Worker
}

func NewClientsCache()(*ClientsCache){
	return &ClientsCache{
		lock: &sync.RWMutex{},
		clients: make(map[string]*Worker),
	}
}

func (c *ClientsCache)Add(identity string, w *Worker){
	c.lock.Lock()
	c.clients[identity] = w
	c.lock.Unlock()
}

func (c *ClientsCache)Delete(identity string){
	c.lock.Lock()
	delete (c.clients, identity)
	c.lock.Unlock()
}

func (c *ClientsCache)GetKeys()([]string){
	keys:=make([]string, 0)

	c.lock.RLock()
	for k := range c.clients{
		keys=append(keys, k)
	}
	c.lock.RUnlock()
	return keys
}

func (c *ClientsCache)GetWorker(identity string)(*Worker, error){

	c.lock.RLock()
	if w,ok:=c.clients[identity]; !ok{
		c.lock.RUnlock()
		return &Worker{}, fmt.Errorf("Client not registered")
	}else{
		c.lock.RUnlock()
		return w, nil
	}
}
