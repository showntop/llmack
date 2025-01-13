package outbound

import (
	"encoding/base64"
	"os"

	"github.com/gorilla/websocket"
	"github.com/showntop/llmack/speech"
)

// WsOutbound ...
type WsOutbound struct {
	*RateLimitOutbound
	ws *websocket.Conn

	ff *os.File
}

// NewWsOutbound ...
func NewWsOutbound(ws *websocket.Conn) speech.Outbound {
	out := &WsOutbound{ws: ws}
	out.RateLimitOutbound = NewRateLimitOutbound()
	out.handle = out.write
	out.ff, _ = os.Create("out-ws.wav")
	return out
}

func (o *WsOutbound) write(data []byte) error {
	rr := base64.StdEncoding.EncodeToString(data)
	o.ff.Write([]byte(rr))
	o.ff.Write([]byte{'\n'})
	// log.InfoContextf(context.TODO(), "message %s, out", rr[:10])
	// return o.ws.WriteMessage(websocket.BinaryMessage, data)
	return o.ws.WriteJSON(map[string]any{
		"type": "websocket_audio",
		"data": rr,
	})
}

// Close ...
func (o *WsOutbound) Close() error {
	return o.ws.Close()
}
