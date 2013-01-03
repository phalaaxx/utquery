utquery
=======

UT2004 server query library for Go language.


Installation
------------
	go get github.com/phalaaxx/utquery


Usage
-----
	package main

	import (
		"fmt"
		"time"
		"github.com/phalaaxx/utquery"
	)                       

	func main() {
		s := utquery.ServerInfo{}
		s.Connect("tauri.deck17.com:7778")
		read := make(chan bool)
		go s.ReceiveData(read)
		timeout := time.After(2*time.Second)
		select {
			case <-timeout:
				fmt.Println("timeout")
			case <-read:
				fmt.Println(s)
		}               
	}               
