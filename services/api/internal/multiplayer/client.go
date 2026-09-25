package multiplayer

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/coder/websocket"
)

// Kode close custom (4000-4999) biar klien bisa bedain alasan putus: yang ini
// jangan di-reconnect otomatis.
const (
	closeReplaced     websocket.StatusCode = 4001
	closeKicked       websocket.StatusCode = 4003
	closeRoomNotFound websocket.StatusCode = 4004
)

const (
	sendBuffer   = 32
	writeTimeout = 10 * time.Second
	// Ping berkala biar koneksi idle (nunggu di lobby) gak diputus proxy
	// (nginx proxy_read_timeout) & koneksi mati kedeteksi.
	pingInterval   = 25 * time.Second
	maxInboundSize = 4 << 10
)

// client: satu koneksi WebSocket. Nulis ke socket cuma dari writeLoop; pihak
// lain (room, di bawah lock) cuma naro pesan ke antrian `send` tanpa blocking,
// jadi klien yang lemot gak nahan room.
type client struct {
	conn *websocket.Conn
	send chan []byte
	quit chan struct{}
	once sync.Once

	closeCode   websocket.StatusCode
	closeReason string
}

func newClient(conn *websocket.Conn) *client {
	return &client{conn: conn, send: make(chan []byte, sendBuffer), quit: make(chan struct{})}
}

func (c *client) sendJSON(v any) {
	b, err := json.Marshal(v)
	if err != nil {
		return
	}
	select {
	case c.send <- b:
	default:
		// Antrian penuh = klien gak kebaca-baca, putus aja (nanti reconnect
		// dan dapet snapshot baru).
		c.shutdown(websocket.StatusPolicyViolation, "slow_consumer")
	}
}

func (c *client) shutdown(code websocket.StatusCode, reason string) {
	c.once.Do(func() {
		c.closeCode = code
		c.closeReason = reason
		close(c.quit)
	})
}

func (c *client) writeLoop(ctx context.Context) {
	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()

	write := func(b []byte) error {
		wctx, cancel := context.WithTimeout(ctx, writeTimeout)
		defer cancel()
		return c.conn.Write(wctx, websocket.MessageText, b)
	}

	for {
		select {
		case b := <-c.send:
			if err := write(b); err != nil {
				c.conn.CloseNow()
				return
			}
		case <-ticker.C:
			pctx, cancel := context.WithTimeout(ctx, writeTimeout)
			err := c.conn.Ping(pctx)
			cancel()
			if err != nil {
				c.conn.CloseNow()
				return
			}
		case <-c.quit:
			// Kirim sisa antrian (mis. pesan "kicked"/"closed") sebelum nutup.
		drain:
			for {
				select {
				case b := <-c.send:
					if write(b) != nil {
						c.conn.CloseNow()
						return
					}
				default:
					break drain
				}
			}
			c.conn.Close(c.closeCode, c.closeReason)
			return
		case <-ctx.Done():
			c.conn.CloseNow()
			return
		}
	}
}

type inbound struct {
	Type   string   `json:"type"`
	Index  int      `json:"index"`
	Value  *float64 `json:"value"`
	UserID string   `json:"userId"`
	Plays  *bool    `json:"plays"`
}

// Serve: jalanin 1 koneksi WebSocket sampai putus. Dipanggil httpserver
// setelah ticket-nya valid & handshake sukses.
func (h *Hub) Serve(ctx context.Context, conn *websocket.Conn, userID, code string) {
	conn.SetReadLimit(maxInboundSize)
	room := h.lookup(code)
	if room == nil {
		conn.Close(closeRoomNotFound, "room_not_found")
		return
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	c := newClient(conn)
	if err := room.attach(userID, c); err != nil {
		conn.Close(closeRoomNotFound, "room_not_found")
		return
	}
	defer room.detach(userID, c)

	go func() {
		c.writeLoop(ctx)
		cancel()
	}()

	for {
		_, data, err := conn.Read(ctx)
		if err != nil {
			return
		}
		var msg inbound
		if json.Unmarshal(data, &msg) != nil {
			continue
		}
		if errCode := room.handle(userID, msg); errCode != "" {
			c.sendJSON(map[string]string{"type": "error", "code": errCode})
		}
	}
}

func (r *Room) handle(userID string, msg inbound) string {
	switch msg.Type {
	case "answer":
		if msg.Value == nil {
			return "invalid_message"
		}
		return r.submitAnswer(userID, msg.Index, *msg.Value)
	case "start":
		return r.start(userID)
	case "kick":
		return r.kick(userID, msg.UserID)
	case "set_host_plays":
		if msg.Plays == nil {
			return "invalid_message"
		}
		return r.setHostPlays(userID, *msg.Plays)
	case "close":
		return r.closeByHost(userID)
	case "reveal_results":
		return r.revealResults(userID)
	case "leave":
		r.leave(userID)
		return ""
	}
	return "invalid_message"
}
