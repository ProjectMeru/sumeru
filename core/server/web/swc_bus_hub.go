package web

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
	"sumeru/core/orm"
	"sumeru/core/queue"
)

var (
	swcBusUpgrader = websocket.Upgrader{
		CheckOrigin: checkSwcBusOrigin,
	}
	globalBusHub     *busHub
	globalBusHubOnce sync.Once
)

func checkSwcBusOrigin(r *http.Request) bool {
	origin := strings.TrimSpace(r.Header.Get("Origin"))
	if origin == "" {
		return true
	}
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	reqHost := r.Host
	if h, _, err := net.SplitHostPort(reqHost); err == nil {
		reqHost = h
	}
	originHost := u.Hostname()
	return strings.EqualFold(originHost, reqHost)
}

type swcBusClient struct {
	uid  int
	conn *websocket.Conn
	send chan []byte
}

type busHub struct {
	mu      sync.RWMutex
	clients map[*swcBusClient]struct{}
}

func ensureBusHub() *busHub {
	globalBusHubOnce.Do(func() {
		globalBusHub = &busHub{clients: make(map[*swcBusClient]struct{})}
		queue.Subscribe("outbox", func(ctx context.Context, msg queue.Message) error {
			var envelope map[string]interface{}
			if err := json.Unmarshal(msg.Payload, &envelope); err != nil {
				return nil
			}
			name, _ := envelope["name"].(string)
			if name == "" {
				return nil
			}
			inner, _ := envelope["payload"].(map[string]interface{})
			if inner == nil {
				inner = map[string]interface{}{}
			}
			actor, _ := orm.CoerceInt64(envelope["actor"])
			out, err := json.Marshal(map[string]interface{}{
				"channel": name,
				"payload": inner,
			})
			if err != nil {
				return nil
			}
			globalBusHub.broadcast(int(actor), out)
			return nil
		})
	})
	return globalBusHub
}

func (h *busHub) register(c *swcBusClient) {
	h.mu.Lock()
	h.clients[c] = struct{}{}
	h.mu.Unlock()
}

func (h *busHub) unregister(c *swcBusClient) {
	h.mu.Lock()
	delete(h.clients, c)
	h.mu.Unlock()
}

func (h *busHub) broadcast(actor int, msg []byte) {
	h.mu.RLock()
	targets := make([]*swcBusClient, 0, len(h.clients))
	for c := range h.clients {
		if actor <= 0 || c.uid == actor {
			targets = append(targets, c)
		}
	}
	h.mu.RUnlock()
	for _, c := range targets {
		select {
		case c.send <- msg:
		default:
		}
	}
}

func (c *swcBusClient) writePump() {
	defer c.conn.Close()
	for msg := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			return
		}
	}
}

func (c *swcBusClient) readPump(h *busHub) {
	defer func() {
		h.unregister(c)
		close(c.send)
		c.conn.Close()
	}()
	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			return
		}
	}
}

func serveSwcBusWebSocket(w http.ResponseWriter, r *http.Request, uid int) {
	hub := ensureBusHub()
	conn, err := swcBusUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	client := &swcBusClient{
		uid:  uid,
		conn: conn,
		send: make(chan []byte, 16),
	}
	hub.register(client)
	go client.writePump()
	client.readPump(hub)
}
