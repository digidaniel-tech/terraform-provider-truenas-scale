package websockethelper

type WebSocketMessage struct {
    ID               string            `json:"id"`
    Name             string            `json:"name"`
    State            string            `json:"state"`
    UpgradeAvailable bool              `json:"upgrade_available"`
    HumanVersion     string            `json:"human_version"`
    Version          string            `json:"version"`
    Metadata         map[string]string `json:"metadata"`
    ActiveWorkloads  map[string]string `json:"active_workloads"`
}
