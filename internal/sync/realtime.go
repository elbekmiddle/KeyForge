package sync

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

// pushMessage mirrors the backend's SyncGateway payload
// ({"type":"profiles.updated","deviceGuid":"..."}).
type pushMessage struct {
	Type       string `json:"type"`
	DeviceGUID string `json:"deviceGuid"`
}

// Listen connects to the backend's real-time /sync channel and calls
// onUpdate whenever the server reports this device's profiles changed
// (doc section 10). It reconnects with backoff on disconnect and blocks
// until ctx is canceled.
func Listen(ctx context.Context, log *slog.Logger, accessToken, guid string, onUpdate func()) {
	backoff := time.Second

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		err := connectAndListen(ctx, accessToken, guid, onUpdate)
		if ctx.Err() != nil {
			return
		}

		if err != nil {
			log.Debug("realtime sync connection lost, retrying", "error", err, "backoff", backoff)
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(backoff):
		}

		if backoff < 30*time.Second {
			backoff *= 2
		}
	}
}

func connectAndListen(ctx context.Context, accessToken, guid string, onUpdate func()) error {
	wsURL, err := toWebSocketURL(BaseURL())
	if err != nil {
		return err
	}

	wsURL.Path = "/sync"

	query := wsURL.Query()
	query.Set("token", accessToken)
	query.Set("deviceGuid", guid)
	wsURL.RawQuery = query.Encode()

	conn, _, err := websocket.DefaultDialer.DialContext(ctx, wsURL.String(), nil)
	if err != nil {
		return fmt.Errorf("sync: failed to connect to realtime channel: %w", err)
	}
	defer conn.Close()

	done := make(chan struct{})
	go func() {
		<-ctx.Done()
		conn.Close()
		close(done)
	}()

	for {
		var msg pushMessage

		if err := conn.ReadJSON(&msg); err != nil {
			select {
			case <-done:
				return nil
			default:
				return fmt.Errorf("sync: realtime channel read failed: %w", err)
			}
		}

		if msg.Type == "profiles.updated" && msg.DeviceGUID == guid {
			onUpdate()
		}
	}
}

func toWebSocketURL(httpURL string) (*url.URL, error) {
	u, err := url.Parse(httpURL)
	if err != nil {
		return nil, fmt.Errorf("sync: invalid backend url %q: %w", httpURL, err)
	}

	switch {
	case strings.HasPrefix(u.Scheme, "https"):
		u.Scheme = "wss"
	default:
		u.Scheme = "ws"
	}

	return u, nil
}
