package server

import (
	"bytes"
	"io"
	"io/fs"
	"net"
	"net/http"
	"net/url"
	"sync"

	_ "embed"

	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header["Origin"]
			if len(origin) == 0 {
				return true
			}
			u, err := url.Parse(origin[0])
			if err != nil {
				return false
			}

			rHost := r.Host
			if forwardedHost := r.Header.Get("X-Forwarded-Host"); forwardedHost != "" {
				rHost = forwardedHost
			}

			if u.Host == rHost {
				return true
			}

			h1, _, err := net.SplitHostPort(u.Host)
			if err != nil {
				return false
			}
			h2, _, err := net.SplitHostPort(r.Host)
			if err != nil {
				return false
			}

			return h1 == h2
		},
		ReadBufferSize: 1024, WriteBufferSize: 1024,
	}
	//go:embed livereload.min.js
	livereloadJS string
)

type client struct {
	ch   chan map[string]any
	conn *websocket.Conn
	once sync.Once
}

func (c *client) loop() {
	for msg := range c.ch {
		err := c.conn.WriteJSON(msg)
		if err != nil {
			break
		}
	}
	c.conn.Close()
}

func (c *client) send(msg map[string]any) {
	c.ch <- msg
}

func (c *client) close() {
	c.once.Do(func() {
		close(c.ch)
	})
}

type Livereload struct {
	mu      sync.Mutex
	clients map[*client]struct{}
}

const injectTag = `<script src="/livereload.js"></script>`

func (m *Livereload) ServeHTML(w http.ResponseWriter, file fs.File) {
	content, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Error reading file", http.StatusInternalServerError)
		return
	}
	// 强行在 </body> 之前注入外链脚本标签
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(bytes.Replace(content, []byte("</body>"), []byte(injectTag+"</body>"), 1))
	return
}

func (m *Livereload) HandleJS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript")
	w.Write([]byte(livereloadJS))
}

func (m *Livereload) HandleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	client := &client{
		ch:   make(chan map[string]any),
		conn: conn,
	}
	go client.loop()

	m.mu.Lock()
	m.clients[client] = struct{}{}
	m.mu.Unlock()

	defer func() {
		m.mu.Lock()
		delete(m.clients, client)
		m.mu.Unlock()

		client.close()
	}()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}
		if bytes.Contains(msg, []byte(`"command":"hello"`)) {
			hello := map[string]any{
				"command":    "hello",
				"protocols":  []string{"http://livereload.com/protocols/official-7"},
				"serverName": "Snow",
			}
			client.send(hello)
		}
	}
}

func (m *Livereload) Notify(path string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	msg := map[string]any{
		"command": "reload",
		"path":    path,
		"liveCSS": true,
	}
	for client := range m.clients {
		client.send(msg)
	}
}
