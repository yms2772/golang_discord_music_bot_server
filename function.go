package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/maxence-charriere/go-app/v7/pkg/app"
	"nhooyr.io/websocket"
)

func (w *web) QueueList() {
	w.queueList = app.Range(w.videoQueueInfo).Slice(func(i int) app.UI {
		return app.Tr().Body(
			app.Td().Body(
				app.Span().Body(
					app.Text(i+1),
				),
			),
			app.Td().Body(
				app.Span().Body(
					app.A().Body(
						app.Text(w.videoQueueInfo[i].Title),
					).
						Href("https://www.youtube.com/watch?v="+w.videoQueueInfo[i].ID).
						Style("color", "white"),
				),
			),
			app.Td().Body(
				app.Span().Body(
					app.Text(w.remainTime[i]),
				),
			),
			app.If(i == 0,
				app.Td().Body(
					app.Span().Body(
						app.A().Body(
							app.Text("건너뛰기"),
						).
							Href("#").
							OnClick(func(ctx app.Context, e app.Event) {
								w.Loading(true)
								_ = w.SendWebsocket(websocket.MessageText, map[string]interface{}{
									"type":          "queue_skip",
									"access_token":  w.accessToken,
									"refresh_token": w.refreshToken,
									"guild_id":      w.videoQueueInfo[i].GuildID,
								})
							}),
					),
				),
			).Else(
				app.Td().Body(
					app.Span().Body(
						app.A().Body(
							app.Text("삭제"),
						).
							Href("#").
							OnClick(func(ctx app.Context, e app.Event) {
								w.Loading(true)
								_ = w.SendWebsocket(websocket.MessageText, map[string]interface{}{
									"type":          "queue_delete",
									"access_token":  w.accessToken,
									"refresh_token": w.refreshToken,
									"guild_id":      w.videoQueueInfo[i].GuildID,
									"queue_id":      w.videoQueueInfo[i].QueueID,
								})
							}),
					),
				),
			),
		)
	})
	w.Update()
}

func (w *web) SendWebsocket(messageType websocket.MessageType, data map[string]interface{}) error {
	if w.conn == nil {
		return errors.New("error")
	}

	ctx := context.Background()
	sendData, _ := json.Marshal(data)

	err := w.conn.Write(ctx, messageType, sendData)
	if err != nil {
		return err
	}

	return nil
}

func (w *web) AddQueue(q string) {
	if len(q) == 0 {
		return
	}

	app.Window().Get("document").Call("getElementById", "searchInput").Set("value", "")

	w.Loading(true)
	w.searchPlaceholder = "검색 중..."
	w.Update()

	_ = w.SendWebsocket(websocket.MessageText, map[string]interface{}{
		"type":          "add_queue",
		"access_token":  w.accessToken,
		"refresh_token": w.refreshToken,
		"search":        q,
	})
}

func (w *web) VerifyUser() bool {
	_ = w.SendWebsocket(websocket.MessageText, map[string]interface{}{
		"type":          "verify_user",
		"access_token":  w.accessToken,
		"refresh_token": w.refreshToken,
	})

	return <-verified
}

func (w *web) AddVerifyUser() {
	rand.Seed(time.Now().UnixNano())

	w.code = strconv.Itoa(rand.Intn(89999) + 10000)
	w.codeHidden = false
	w.Update()

	_ = w.SendWebsocket(websocket.MessageText, map[string]interface{}{
		"type":          "add_verify_user",
		"access_token":  w.accessToken,
		"refresh_token": w.refreshToken,
		"guild_id":      w.guildID,
		"ip":            w.ip,
		"code":          w.code,
	})

	go func() {
		if <-addVerify {
			app.Window().Call("alert", "인증 완료")

			_ = w.SendWebsocket(websocket.MessageText, map[string]interface{}{
				"type":          "queue_list",
				"access_token":  w.accessToken,
				"refresh_token": w.refreshToken,
			})
			w.Loading(false)
		} else {
			app.Window().Call("alert", "인증 실패\n\n새로고침 후 다시 시도해주세요")
		}

		w.codeHidden = true
		w.Update()
	}()
}

func (w *web) Loading(ok bool, msg ...string) {
	if ok {
		if len(msg) == 0 {
			w.loadingText = "loading"
		} else {
			w.loadingText = strings.Join(msg, " ")
		}

		w.pointerEventMain = "none"
		w.blurMain = "blur(10px)"
		w.loadingHidden = false
	} else {
		w.pointerEventMain = "all"
		w.blurMain = "blur(0px)"
		w.loadingHidden = true
	}

	w.Update()
}

func (w *web) ChannelJoinStatus() bool {
	_ = w.SendWebsocket(websocket.MessageText, map[string]interface{}{
		"type":          "channel_join_status",
		"access_token":  w.accessToken,
		"refresh_token": w.refreshToken,
		"guild_id":      w.guildID,
	})

	return <-channelJoinStatus
}

func (w *web) EncryptUniq(guildID, ip string) string {
	hash := sha256.New()
	hash.Write([]byte(guildID + ip))

	return hex.EncodeToString(hash.Sum(nil))
}

func (w *web) GetAccessToken(uniq string) string {
	hash := sha256.New()
	hash.Write([]byte(fmt.Sprintf("%s%d", uniq, time.Now().Day())))

	return hex.EncodeToString(hash.Sum(nil))
}
