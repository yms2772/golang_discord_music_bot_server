package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/maxence-charriere/go-app/v7/pkg/app"
	"nhooyr.io/websocket"
)

func (w *web) OnMount(ctx app.Context) {
	var err error

	urlPaths := strings.Split(app.Window().URL().Path, "/")

	if len(urlPaths) < 3 {
		log.Println(urlPaths)
		return
	}

	w.remainTime = make(map[int]string)
	w.guildID = urlPaths[2]
	w.ip = app.Window().Call("ip").String()
	w.refreshToken = w.EncryptUniq(w.guildID, w.ip)
	w.accessToken = w.GetAccessToken(w.refreshToken)
	w.pointerEventMain = "none"
	w.blurMain = "blur(10px)"
	w.loadingHidden = true
	w.codeHidden = true
	w.searchPlaceholder = "대기열 추가"
	w.play = "play-play"
	w.Update()

	if len(w.guildID) == 0 {
		return
	}

	w.conn, _, err = websocket.Dial(ctx, WebsocketAddress+"/"+w.guildID, nil)
	if err != nil {
		return
	}

	go func() {
		defer w.conn.Close(websocket.StatusInternalError, "StatusInternalError")

		for {
			_, message, err := w.conn.Read(ctx)
			if err != nil {
				app.Reload()
			}

			var receive Receive
			json.Unmarshal(message, &receive)

			switch receive.Type {
			case "alert":
				app.Window().Call("alert", receive.Msg)
			case "token_refresh_required":
				w.accessToken = w.GetAccessToken(w.refreshToken)
			case "token_expired":
				app.Reload()
			case "require_verify":
				w.AddVerifyUser()
			case "verify_done":
				w.Loading(false)
			case "verify":
				if receive.Verify {
					w.userID = receive.UserID
				}

				verified <- receive.Verify
			case "add_verify":
				log.Println("add_verify")
				addVerify <- receive.Verify
				log.Println("done: add_verify")
			case "channel_join_status":
				channelJoinStatus <- receive.ChannelJoinStatus
			case "loading_done":
				w.Loading(false)
			case "add_queue_done":
				w.searchPlaceholder = "대기열 추가"
				w.Update()
			case "update_queue":
				w.videoQueueInfo = receive.VideoQueueInfo

				if len(w.videoQueueInfo) == 0 {
					w.thumbnail = ""
				} else {
					w.thumbnail = w.videoQueueInfo[0].Thumbnail
				}

				w.QueueList()
			case "update_time":
				w.progress = (receive.PlaybackPosition / float64(receive.Duration)) * 100

				if len(w.videoQueueInfo) != 0 {
					videoDurationSeconds := w.videoQueueInfo[0].Duration
					videoCurrentSeconds := int(receive.PlaybackPosition)

					videoCurrentH := videoCurrentSeconds / 3600
					videoCurrentM := (videoCurrentSeconds - (3600 * videoCurrentH)) / 60
					videoCurrentS := videoCurrentSeconds - (3600 * videoCurrentH) - (videoCurrentM * 60)

					videoDurationH := videoDurationSeconds / 3600
					videoDurationM := (videoDurationSeconds - (3600 * videoDurationH)) / 60
					videoDurationS := videoDurationSeconds - (3600 * videoDurationH) - (videoDurationM * 60)

					w.timeline = fmt.Sprintf("%02d:%02d:%02d / %02d:%02d:%02d", videoCurrentH, videoCurrentM, videoCurrentS, videoDurationH, videoDurationM, videoDurationS)
				}

				w.Update()

				var videoRemainSeconds int

				for i := range w.videoQueueInfo {
					if i == 0 {
						w.remainTime[i] = "현재 재생 중"
					} else {
						if i-1 == 0 {
							videoRemainSeconds += w.videoQueueInfo[i-1].Duration - int(receive.PlaybackPosition)
						} else {
							videoRemainSeconds += w.videoQueueInfo[i-1].Duration
						}

						videoRemainH := videoRemainSeconds / 3600
						videoRemainM := (videoRemainSeconds - (3600 * videoRemainH)) / 60
						videoRemainS := videoRemainSeconds - (3600 * videoRemainH) - (videoRemainM * 60)

						w.remainTime[i] = fmt.Sprintf("%02d:%02d:%02d", videoRemainH, videoRemainM, videoRemainS)
					}
				}

				w.QueueList()
			}
		}
	}()
}

func (w *web) OnNav(_ app.Context, _ *url.URL) {
	w.Loading(true, "인증 중")
	_ = w.SendWebsocket(websocket.MessageText, map[string]interface{}{
		"type":          "verify",
		"access_token":  w.accessToken,
		"refresh_token": w.refreshToken,
	})
	_ = w.SendWebsocket(websocket.MessageText, map[string]interface{}{
		"type":          "queue_list",
		"access_token":  w.accessToken,
		"refresh_token": w.refreshToken,
	})
}

func (w *web) Render() app.UI {
	return app.Div().Body(
		app.Div().Body(
			app.Span().Body(
				app.Text("인증번호: "+w.code),
			).
				Style("color", "white").
				Style("font-size", "50px").
				Style("position", "absolute").
				Hidden(w.codeHidden),
		),
		app.Div().Body(
			app.Div().Body().
				Class("loading"),
			app.Div().Body(
				app.Text(w.loadingText),
			).
				ID("loading-text"),
		).
			Class("loading-container").
			Hidden(w.loadingHidden),
		app.Div().Body(
			app.If(len(w.videoQueueInfo) != 0,
				app.Div().Body(
					app.Div().Body(
						app.Img().
							Src(w.thumbnail).
							Style("width", "100%").
							Style("height", "100%"),
					).
						Class("thumbnail"),
					app.Div().Body(
						app.Div().Body(
							app.Div().Body().
								Class("bar").
								ID("audio-progress-bar").
								Style("width", fmt.Sprintf("%.2f%%", w.progress)),
						).
							Class("audio-progress").
							ID("audio-progress").
							OnClick(func(ctx app.Context, e app.Event) {
								if len(w.videoQueueInfo) != 0 {
									w.Loading(true)

									offsetX := (float64(e.Get("offsetX").Int()) / float64(app.Window().Get("document").Call("getElementById", "audio-progress").Get("offsetWidth").Int())) * 100

									_ = w.SendWebsocket(websocket.MessageText, map[string]interface{}{
										"type":          "play_jump",
										"access_token":  w.accessToken,
										"refresh_token": w.refreshToken,
										"start_time":    int((offsetX / 100) * float64(w.videoQueueInfo[0].Duration)),
									})
								}
							}),
						app.Div().Body(
							app.Text(w.timeline),
						).
							Class("timeline"),
					).
						ID("audio-player-container"),
					app.Div().Body(
						app.Span().Body().
							Class(w.play).
							Title("재생/정지").
							OnClick(func(ctx app.Context, e app.Event) {
								if w.play == "play-play" {
									w.play = "play-pause"
								} else {
									w.play = "play-play"
								}

								w.Update()

								_ = w.SendWebsocket(websocket.MessageText, map[string]interface{}{
									"type":          "play_pause",
									"access_token":  w.accessToken,
									"refresh_token": w.refreshToken,
								})
							}),
					),
				).
					Class("left"),
			),
			app.Div().Body(
				app.Div().Body(
					app.Div().Body(
						app.Table().Body(
							app.THead().Body(
								app.Tr().Body(
									app.Th().Body(
										app.Text("순서"),
									),
									app.Th().Body(
										app.Text("제목"),
									),
									app.Th().Body(
										app.Text("남은 시간"),
									),
									app.Th().Body(
										app.Text("작업"),
									),
								),
							),
							app.TBody().Body(
								w.queueList,
							),
							app.Div().Body(
								app.Link().
									Rel("stylesheet").
									Href("https://fonts.googleapis.com/icon?family=Material+Icons"),
								app.Input().
									Class("searchInput").
									ID("searchInput").
									Type("text").
									Placeholder(w.searchPlaceholder).
									OnKeyPress(func(ctx app.Context, e app.Event) {
										if e.Get("keyCode").Int() == 13 {
											w.AddQueue(ctx.JSSrc.Get("value").String())
										}
									}),
								app.Button().Body(
									app.I().Body(
										app.Text("search"),
									).
										Class("material-icons"),
								).
									Class("searchButton").
									OnClick(func(ctx app.Context, e app.Event) {
										w.AddQueue(app.Window().Get("document").Call("getElementById", "searchInput").Get("value").String())
									}),
							).
								Class("searchBox"),
						).
							Class("design2-table"),
					).
						Class("design2-mypage-content"),
				).
					Class("design2-content-box"),
			).
				Class("right"),
		).
			Class("main").
			Style("pointer-events", w.pointerEventMain).
			Style("filter", w.blurMain),
	)
}
