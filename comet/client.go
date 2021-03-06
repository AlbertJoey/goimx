package comet

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/AlbertJoey/goimx/model"
	"github.com/gorilla/websocket"
)

type client struct {
	comet  *comet
	conn   *websocket.Conn
	id     int64
	tag    string
	ch     chan *model.DTO
	stop   chan error
	groups []int64
	rw     sync.RWMutex
}

func new(comet *comet, conn *websocket.Conn, id int64, tag string) *client {
	return &client{
		comet:  comet,
		conn:   conn,
		id:     id,
		tag:    tag,
		ch:     make(chan *model.DTO, 0),
		stop:   make(chan error),
		groups: make([]int64, 0),
		rw:     sync.RWMutex{},
	}
}

func (cli *client) run() {
	go cli.recv()
	go cli.send()
	<-cli.stop
}

//recv cli msg
func (cli *client) recv() {
	for {
		j := &model.DTO{}
		_, message, err := cli.conn.ReadMessage()
		if err != nil {
			cli.quitallgrp()
			cli.comet.mvcli(cli)
			return
		}
		if j.Msg == nil {
			j.Msg = &model.Msg{}
		}
		j.Msg.FromUserID = cli.id
		j.Msg.FromUserTag = cli.tag
		// fmt.Println("^^^^^^^^^ Comet Client Recv:", string(message))
		err = json.Unmarshal(message, j)
		if err == nil {
			if j.Type == model.CliJoinGrp { //加群
				cli.joingrp(j)
			}
			if j.Type == model.CliQuitGrp { //退群
				cli.quitgrp(j)
			}
			ch := cli.comet.getch()
			ch <- j
		}
	}
}

//send msg to cli
func (cli *client) send() {
	for {
		select {
		case dto := <-cli.ch:
			{
				j, err := json.Marshal(dto)
				if err == nil {
					err = cli.conn.WriteMessage(websocket.TextMessage, j)
					if err != nil {
						cli.quitallgrp()
						cli.comet.mvcli(cli)
						return
					}
				}
			}
		}
	}
}

//监听
func Serve(w http.ResponseWriter, r *http.Request, c *comet) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("cli upgrade:", err)
		return
	}
	defer conn.Close()
	//get token and tag
	query := r.URL.Query()
	token := query.Get("token")
	tag := query.Get("tag")
	id, err := strconv.ParseInt(token, 10, 64)
	if err != nil || tag == "" {
		return
	}
	cli := new(c, conn, id, tag)
	if cli.comet.isexist(cli) {
		return
	}
	cli.comet.addcli(cli)
	cli.run()
}

//退出指定群
func (cli *client) quitgrp(dto *model.DTO) {
	cli.rw.Lock()
	idx, isexist := 0, false
	for i, grpid := range cli.groups {
		if dto.Msg.GroupID == grpid {
			idx, isexist = i, true
			cli.comet.grpmvcli(cli, grpid)
		}
	}
	if isexist {
		cli.groups = append(cli.groups[:idx], cli.groups[idx+1:]...)
	}
	cli.rw.Unlock()
}

//退出所有群
func (cli *client) quitallgrp() {
	cli.rw.Lock()
	for _, grpid := range cli.groups {
		cli.comet.grpmvcli(cli, grpid)
	}
	cli.groups = []int64{}
	cli.rw.Unlock()
}

//加群
func (cli *client) joingrp(dto *model.DTO) {
	cli.rw.Lock()
	isexist := false
	for _, grpid := range cli.groups {
		if grpid == dto.Msg.GroupID {
			isexist = true
		}
	}
	if !isexist {
		cli.groups = append(cli.groups, dto.Msg.GroupID)
	}
	cli.comet.grpaddcli(cli, dto.Msg.GroupID)
	cli.rw.Unlock()
}
