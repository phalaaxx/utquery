// Copyright 2012-2020 Bozhin Zafirov. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/* Package utquery is a UT2004 query library for Go */
package utquery

import (
	"bytes"
	"encoding/binary"
	"io"
	"net"
	"strings"
	"time"
)

/* constants */
const QRY_SERVERINFO = "\x80\x00\x00\x00\x00"
const QRY_GAMEINFO = "\x80\x00\x00\x00\x01"
const QRY_PLAYERSINFO = "\x80\x00\x00\x00\x02"

/* data parser */
type Buffer struct {
	data *bytes.Buffer
}

/* read data from socket */
func (b *Buffer) ReceiveData(conn net.Conn) error {
	buffer := make([]byte, 2048)
	if n, err := conn.Read(buffer); err != nil {
		return err
	} else {
		b.data = bytes.NewBuffer(buffer[0:n])
	}
	return nil
}

/* parse integer from buffer */
func (b *Buffer) GetInt() (ret int32) {
	binary.Read(b.data, binary.LittleEndian, &ret)
	return ret
}

/* parse string from buffer */
func (b *Buffer) GetString() (ret string) {
	length, err := b.data.ReadByte()
	if err != nil || length == 0 {
		return ret
	}
	StrData := make([]byte, length)
	if _, err := io.ReadFull(b.data, StrData); err != nil {
		return ret
	}
	if bytes.Contains(StrData, []byte("\x1b\n\xf5\n")) {
		StrData = StrData[4:]
	}
	return string(StrData)
}

/* true if there is more data to parse */
func (b *Buffer) HasData() bool {
	return b.data.Len() != 0
}

/* Player Info structure */
type PlayersInfo struct {
	ID      int32
	Name    string
	Ping    int32
	Score   int32
	StatsID int32
}

/* Server Info structure */
type ServerInfo struct {
	ID         int32
	IP         string
	Address    string
	Port       int32
	SQPort     int32
	Name       string
	Map        string
	MapImage   string
	GameType   string
	Players    int32
	MaxPlayers int32
	Ping       int32
	Flags      int32
	SkillLevel int32

	/* array of players */
	PlayersList []PlayersInfo
	GameInfo    map[string]string

	/* private data */
	players int
	conn    net.Conn
}

/* initialize connection to remote server */
func (q *ServerInfo) Connect(serverAddr string) error {
	q.Address = serverAddr[0:strings.Index(serverAddr, ":")]
	if conn, err := net.Dial("udp", serverAddr); err != nil {
		return err
	} else {
		conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		q.conn = conn
	}
	q.conn.Write([]byte(QRY_SERVERINFO))
	q.conn.Write([]byte(QRY_GAMEINFO))
	q.conn.Write([]byte(QRY_PLAYERSINFO))

	return nil
}

/* receive and parse data */
func (q *ServerInfo) ReceiveData(read chan bool) {
	for i := 0; i < 3; i++ {
		b := new(Buffer)
		if err := b.ReceiveData(q.conn); err != nil {
			read <- false
		}
		if bytes.Contains(b.data.Bytes()[0:], []byte(QRY_SERVERINFO)) {
			b.data.Next(len(QRY_SERVERINFO))
			q.ID = b.GetInt()
			q.IP = b.GetString()
			q.Port = b.GetInt()
			q.SQPort = b.GetInt()
			q.Name = b.GetString()
			q.Map = b.GetString()
			q.GameType = b.GetString()
			q.Players = b.GetInt()
			q.MaxPlayers = b.GetInt()
			q.Ping = b.GetInt()
			q.Flags = b.GetInt()
			q.SkillLevel = b.GetInt()
		}
		if bytes.Contains(b.data.Bytes()[0:], []byte(QRY_GAMEINFO)) {
			b.data.Next(len(QRY_GAMEINFO))
			q.GameInfo = make(map[string]string)
			for b.HasData() {
				q.GameInfo[b.GetString()] = b.GetString()
			}
		}
		if bytes.Contains(b.data.Bytes()[0:], []byte(QRY_PLAYERSINFO)) {
			b.data.Next(len(QRY_PLAYERSINFO))
			for b.HasData() {
				p := &PlayersInfo{
					ID:      b.GetInt(),
					Name:    b.GetString(),
					Ping:    b.GetInt(),
					Score:   b.GetInt(),
					StatsID: b.GetInt(),
				}
				q.PlayersList = append(q.PlayersList, *p)
			}
		}
	}
	read <- true
}
