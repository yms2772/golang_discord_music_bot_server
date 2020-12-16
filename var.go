package main

const (
	websocketAddress = ""
)

var (
	verified          = make(chan bool)
	addVerify         = make(chan bool)
	channelJoinStatus = make(chan bool)
)
