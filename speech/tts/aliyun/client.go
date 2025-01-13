package aliyun

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WsClient ...
type WsClient struct {
	sync.Mutex
	ws *websocket.Conn
}

// NewWsClient ...
func NewWsClient(token string) (*WsClient, error) {
	wsc := &WsClient{}

	// url := "wss://nls-gateway-cn-beijing.aliyuncs.com/ws/v1?token=%s"
	// url = fmt.Sprintf(url, token)

	// dialer := websocket.Dialer{}
	// header := http.Header(make(map[string][]string))
	// conn, _, err := dialer.Dial(url, header)
	// if err != nil {
	// 	return nil, err
	// }
	// wsc.ws = conn
	// fmt.Println("new")

	// go wsc.keepAlive()

	return wsc, nil
}

// reconnect ...
func (c *WsClient) reconnect(token string) error {
	fmt.Println("reconnecting 1")

	c.Lock()
	defer c.Unlock()

	url := "wss://nls-gateway-cn-beijing.aliyuncs.com/ws/v1?token=%s"
	url = fmt.Sprintf(url, token)
	fmt.Println("reconnecting")
	dialer := websocket.Dialer{}
	header := http.Header(make(map[string][]string))
	conn, _, err := dialer.Dial(url, header)
	if err != nil {
		return err
	}
	fmt.Println("reconnecting done")
	c.ws = conn

	return nil
}

// WriteJSON ...
func (c *WsClient) WriteJSON(x any) error {
	c.Lock()
	defer c.Unlock()

	return c.ws.WriteJSON(x)
}

func (c *WsClient) keepAlive() {
	for {
		fmt.Println("keepAlive")
		c.Lock()
		if err := c.ws.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(10*time.Second)); err != nil {
			panic(err)
		}
		c.Unlock()

		time.Sleep(9 * time.Second)
	}
}

// Close ...
func (c *WsClient) close() error {
	c.Lock()
	defer c.Unlock()

	return c.ws.Close()
}

// // Close ...
// func (c *WsClient) Write() error {
// 	return c.ws.Close()
// }
