package websockethelper

type JobResponse struct {
    Msg    string               `json:"msg"`
	ID     string               `json:"id"`
	Result []JobResponseResult  `json:"result"`
}

type JobResponseResult struct {
	ID       int                    `json:"id"`
	State    string                 `json:"state"`
	Progress JobResponseProgress    `json:"progress"`
	Error  interface{}              `json:"error"`
	Result interface{}              `json:"result"`
}

type JobResponseProgress struct {
    Percent     float64 `json:"percent"`
	Description string  `json:"description"`
}
