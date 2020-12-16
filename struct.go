package main

import (
	"github.com/maxence-charriere/go-app/v7/pkg/app"
	"nhooyr.io/websocket"
)

type web struct {
	app.Compo

	queueList app.RangeLoop

	conn              *websocket.Conn
	ip                string
	guildID           string
	userID            string
	currentTime       float64
	remainTime        map[int]string
	videoQueueInfo    []*VideoQueue
	searchPlaceholder string
	progress          float64
	thumbnail         string
	pointerEventMain  string
	loadingHidden     bool
	codeHidden        bool
	blurMain          string
	timeline          string
	play              string
	loadingText       string
	accessToken       string
	refreshToken      string
	code              string
}

type VideoQueue struct {
	QueueID   int
	GuildID   string
	ID        string
	Title     string
	Channel   string
	Duration  int
	Thumbnail string
}

type VideoQueueJSON struct {
	QueueID   int    `json:"queue_id"`
	GuildID   string `json:"guild_id"`
	UserID    string `json:"user_id"`
	ID        string `json:"id"`
	Title     string `json:"title"`
	Channel   string `json:"channel"`
	Duration  int    `json:"duration"`
	Thumbnail string `json:"thumbnail"`
}

type Receive struct {
	VideoQueueJSON

	Type              string        `json:"type"`
	VideoQueueInfo    []*VideoQueue `json:"video_queue_info"`
	PlaybackPosition  float64       `json:"playback_position"`
	Msg               string        `json:"msg"`
	Verify            bool          `json:"verify"`
	ChannelJoinStatus bool          `json:"channel_join_status"`
}
