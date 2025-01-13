// Copyright (c) HashiCorp, Inc.

package websockethelper

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type WebSocketClient struct {
	conn *websocket.Conn
}

func NewWebSocketClient(url string) (*WebSocketClient, error) {
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to WebSocket: %w", err)
	}

	return &WebSocketClient{conn: conn}, nil
}

func (c *WebSocketClient) Send(method string, params interface{}) (string, error) {
	requestID := uuid.NewString()
	message := map[string]interface{}{
		"id":     requestID,
		"msg":    "method",
		"method": method,
		"params": params,
	}

	if err := c.conn.WriteJSON(message); err != nil {
		return "", fmt.Errorf("failed to send message: %w", err)
	}

	return requestID, nil
}

func (c *WebSocketClient) Receive() (map[string]interface{}, error) {
	var response map[string]interface{}
	if err := c.conn.ReadJSON(&response); err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return response, nil
}

func (c *WebSocketClient) PollJobStatus(jobID int) (string, error) {
	for {
		time.Sleep(2 * time.Second)

		queryID := uuid.NewString()
		queryMessage := map[string]interface{}{
			"id":     queryID,
			"msg":    "method",
			"method": "core.job.query",
			"params": [][]interface{}{
				{"id", "=", jobID},
			},
		}

		if err := c.conn.WriteJSON(queryMessage); err != nil {
			return "", fmt.Errorf("failed to send job query: %w", err)
		}

		var response struct {
			Result []struct {
				State string `json:"state"`
				Error string `json:"error"`
			} `json:"result"`
		}

		if err := c.conn.ReadJSON(&response); err != nil {
			return "", fmt.Errorf("failed to read job status: %w", err)
		}

		if len(response.Result) > 0 {
			job := response.Result[0]
			switch job.State {
			case "SUCCESS":
				return "SUCCESS", nil
			case "FAILED":
				return "FAILED", fmt.Errorf("job failed: %s", job.Error)
			default:
				// Job still in progress
			}
		}
	}
}

func (c *WebSocketClient) Close() {
	c.conn.Close()
}
