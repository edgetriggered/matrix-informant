package informant

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/rs/zerolog"

	_ "github.com/mattn/go-sqlite3"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/crypto/cryptohelper"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

type Intelligence struct {
	Channel      string
	Message      string
	ContentBytes []byte
	ContentType  string
	Caption      string
	PSK          string
}

func Inform(config string) {
	c, err := ReadConfig(config)
	if err != nil {
		log.Fatalln(err)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	client, err := mautrix.NewClient(c.Homeserver, "", "")
	if err != nil {
		log.Fatalln(err)
	}

	log := zerolog.New(zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
		w.Out = os.Stdout
		w.TimeFormat = time.Stamp
	})).With().Timestamp().Logger()
	if !c.Debug {
		log = log.Level(zerolog.InfoLevel)
	}
	client.Log = log

	syncer := client.Syncer.(*mautrix.DefaultSyncer)
	syncer.OnEventType(event.EventMessage, func(source mautrix.EventSource, evt *event.Event) {
		log.Debug().
			Str("sender", evt.Sender.String()).
			Str("type", evt.Type.String()).
			Str("id", evt.ID.String()).
			Str("body", evt.Content.AsMessage().Body).
			Msg("Received message")
	})
	syncer.OnEventType(event.StateMember, func(source mautrix.EventSource, evt *event.Event) {
		if evt.GetStateKey() == client.UserID.String() && evt.Content.AsMember().Membership == event.MembershipInvite {
			_, err := client.JoinRoomByID(evt.RoomID)
			if err == nil {
				log.Info().
					Str("room_id", evt.RoomID.String()).
					Str("inviter", evt.Sender.String()).
					Msg("Joined room after invite")
			} else {
				log.Error().Err(err).
					Str("room_id", evt.RoomID.String()).
					Str("inviter", evt.Sender.String()).
					Msg("Failed to join room after invite")
			}
		}
	})

	cryptoHelper, err := cryptohelper.NewCryptoHelper(client, []byte(c.Database.Key), c.Database.Path)
	if err != nil {
		panic(err)
	}

	cryptoHelper.LoginAs = &mautrix.ReqLogin{
		Type:       mautrix.AuthTypePassword,
		Identifier: mautrix.UserIdentifier{Type: mautrix.IdentifierTypeUser, User: c.Username},
		Password:   c.Password,
	}

	err = cryptoHelper.Init()
	if err != nil {
		panic(err)
	}

	client.Crypto = cryptoHelper

	syncCtx, cancelSync := context.WithCancel(context.Background())
	var syncStopWait sync.WaitGroup
	syncStopWait.Add(1)

	go func() {
		err = client.SyncWithContext(syncCtx)
		defer syncStopWait.Done()
		if err != nil && !errors.Is(err, context.Canceled) {
			panic(err)
		}
	}()

	ch := make(chan Intelligence)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var data []byte
		var err error

		if r.Body != nil {
			data, err = io.ReadAll(r.Body)
			if err != nil {
				log.Warn().Err(err).Msg("Failed to read POST data")
				return
			}
			defer r.Body.Close()
		}

		i := Intelligence{}
		err = json.Unmarshal(data, &i)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to unmarshal POST data")
			return
		}

		ch <- i
	})

	go func() {
		log.Info().Msg("Listening on " + c.Bind)
		http.ListenAndServe(c.Bind, nil)
	}()

	avatar, err := os.ReadFile(c.Avatar)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read avatar image")
	} else {
		content, err := client.UploadBytes(avatar, "image/png")
		if err != nil {
			log.Error().Err(err).Msg("Failed to upload data")
		} else {
			err = client.SetAvatarURL(content.ContentURI)
			if err != nil {
				log.Error().Err(err).Msg("Failed to set avatar")
			}
		}
	}

	err = client.SetDisplayName(c.Display)
	if err != nil {
		log.Error().Err(err).Msg("Failed to set display name")
	}

	for {
		if <-sig == os.Interrupt {
			break
		}

		i := <-ch

		if bytes.Equal([]byte(i.PSK), []byte(c.PSK)) {
			if !bytes.Equal(i.ContentBytes, []byte("")) {
				upload, err := client.UploadBytes(i.ContentBytes, i.ContentType)
				if err != nil {
					log.Error().Err(err).Msg("Failed to upload data")
				}
				content := event.MessageEventContent{
					MsgType: event.MsgImage,
					Body:    i.Caption,
					URL:     upload.ContentURI.CUString(),
				}
				resp, err := client.SendMessageEvent(id.RoomID(i.Channel), event.EventMessage, &content)
				if err != nil {
					log.Error().Err(err).Msg("Failed to send event")
				} else {
					log.Info().Str("event_id", resp.EventID.String()).Msg("Sent event!")
				}
			}
			if i.Message != "" {
				resp, err := client.SendText(id.RoomID(i.Channel), i.Message)
				if err != nil {
					log.Error().Err(err).Msg("Failed to send message")
				} else {
					log.Info().Str("event_id", resp.EventID.String()).Msg("Sent message!")
				}
			}
		}

		time.Sleep(250 * time.Millisecond)
	}

	log.Warn().Msg("Caught SIGTERM, exiting")

	cancelSync()
	syncStopWait.Wait()

	err = cryptoHelper.Close()
	if err != nil {
		log.Error().Err(err).Msg("Error closing database")
	}
}
