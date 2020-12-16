package main

const (
	WebsocketAddress = ""
)

var (
	verified          = make(chan bool)
	addVerify         = make(chan bool)
	channelJoinStatus = make(chan bool)
)
