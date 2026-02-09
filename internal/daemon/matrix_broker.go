package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/alexcabrera/ayo/internal/paths"
)

// MatrixBroker manages Matrix communication for agents.
// It maintains a single sync connection to Conduit and routes messages.
type MatrixBroker struct {
	socketPath string
	serverName string
	httpClient *http.Client

	mu           sync.RWMutex
	connected    bool
	syncing      bool
	daemonUserID string
	accessToken  string

	// Agent management
	agents       map[string]*AgentUser
	agentClients map[string]*http.Client

	// Room tracking
	rooms map[string]*RoomInfo

	// Message queues
	queues map[string]chan *QueuedMessage

	// Callbacks
	onMessage func(agent string, msg *QueuedMessage)

	// Sync control
	syncCancel context.CancelFunc
}

// AgentUser represents a Matrix user for an agent.
type AgentUser struct {
	UserID   string `json:"user_id"`
	Handle   string `json:"handle"`
	Password string `json:"password"`
	Token    string `json:"token"`
}

// RoomInfo represents a Matrix room.
type RoomInfo struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Topic    string   `json:"topic,omitempty"`
	Session  string   `json:"session,omitempty"`
	Members  []string `json:"members"`
	Unread   int      `json:"unread"`
	LastSync string   `json:"last_sync,omitempty"`
}

// QueuedMessage represents a message waiting to be delivered.
type QueuedMessage struct {
	RoomID    string    `json:"room_id"`
	EventID   string    `json:"event_id"`
	Sender    string    `json:"sender"`
	Handle    string    `json:"handle,omitempty"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	IsReply   bool      `json:"is_reply,omitempty"`
	ReplyTo   string    `json:"reply_to,omitempty"`
}

// BrokerStatus represents the status of the Matrix broker.
type BrokerStatus struct {
	Connected   bool   `json:"connected"`
	Syncing     bool   `json:"syncing"`
	UserID      string `json:"user_id"`
	RoomCount   int    `json:"room_count"`
	AgentCount  int    `json:"agent_count"`
	QueuedMsgs  int    `json:"queued_messages"`
	ServerName  string `json:"server_name"`
}

// NewMatrixBroker creates a new Matrix broker.
func NewMatrixBroker(socketPath, serverName string) *MatrixBroker {
	transport := &http.Transport{
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			return net.Dial("unix", socketPath)
		},
	}

	return &MatrixBroker{
		socketPath:   socketPath,
		serverName:   serverName,
		httpClient:   &http.Client{Transport: transport, Timeout: 30 * time.Second},
		agents:       make(map[string]*AgentUser),
		agentClients: make(map[string]*http.Client),
		rooms:        make(map[string]*RoomInfo),
		queues:       make(map[string]chan *QueuedMessage),
	}
}

// Connect establishes connection to the Matrix server.
func (b *MatrixBroker) Connect(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.connected {
		return nil
	}

	// Load agent credentials
	if err := b.loadAgentCredentials(); err != nil {
		// Not fatal - will register as needed
	}

	// Register or login as daemon user
	daemonHandle := "ayo"
	user, err := b.ensureAgentUserLocked(daemonHandle)
	if err != nil {
		return fmt.Errorf("setup daemon user: %w", err)
	}

	b.daemonUserID = user.UserID
	b.accessToken = user.Token
	b.connected = true

	// Start sync loop
	syncCtx, cancel := context.WithCancel(ctx)
	b.syncCancel = cancel
	go b.syncLoop(syncCtx)

	return nil
}

// Disconnect closes the Matrix connection.
func (b *MatrixBroker) Disconnect() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.syncCancel != nil {
		b.syncCancel()
	}

	b.connected = false
	b.syncing = false

	// Save agent credentials
	b.saveAgentCredentials()

	return nil
}

// Connected returns true if connected to Matrix.
func (b *MatrixBroker) Connected() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.connected
}

// Status returns the broker status.
func (b *MatrixBroker) Status() BrokerStatus {
	b.mu.RLock()
	defer b.mu.RUnlock()

	totalQueued := 0
	for _, q := range b.queues {
		totalQueued += len(q)
	}

	return BrokerStatus{
		Connected:   b.connected,
		Syncing:     b.syncing,
		UserID:      b.daemonUserID,
		RoomCount:   len(b.rooms),
		AgentCount:  len(b.agents),
		QueuedMsgs:  totalQueued,
		ServerName:  b.serverName,
	}
}

// SetOnMessage sets the callback for incoming messages.
func (b *MatrixBroker) SetOnMessage(handler func(agent string, msg *QueuedMessage)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.onMessage = handler
}

// EnsureAgentUser creates or retrieves a Matrix user for an agent.
func (b *MatrixBroker) EnsureAgentUser(handle string) (*AgentUser, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.ensureAgentUserLocked(handle)
}

func (b *MatrixBroker) ensureAgentUserLocked(handle string) (*AgentUser, error) {
	// Check if already registered
	if user, ok := b.agents[handle]; ok {
		return user, nil
	}

	// Generate credentials
	username := handle
	password := generateRandomString(32)
	userID := fmt.Sprintf("@%s:%s", username, b.serverName)

	// Try to register
	regReq := map[string]interface{}{
		"username": username,
		"password": password,
		"auth": map[string]string{
			"type": "m.login.dummy",
		},
	}

	regBody, _ := json.Marshal(regReq)
	resp, err := b.httpClient.Post(
		"http://localhost/_matrix/client/v3/register",
		"application/json",
		strings.NewReader(string(regBody)),
	)
	if err != nil {
		return nil, fmt.Errorf("register request: %w", err)
	}
	defer resp.Body.Close()

	var regResp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&regResp)

	var token string
	if t, ok := regResp["access_token"].(string); ok {
		token = t
	} else {
		// Registration failed, try login
		token, err = b.loginUser(username, password)
		if err != nil {
			return nil, fmt.Errorf("login failed: %w", err)
		}
	}

	user := &AgentUser{
		UserID:   userID,
		Handle:   handle,
		Password: password,
		Token:    token,
	}
	b.agents[handle] = user

	return user, nil
}

func (b *MatrixBroker) loginUser(username, password string) (string, error) {
	loginReq := map[string]interface{}{
		"type":     "m.login.password",
		"user":     username,
		"password": password,
	}

	loginBody, _ := json.Marshal(loginReq)
	resp, err := b.httpClient.Post(
		"http://localhost/_matrix/client/v3/login",
		"application/json",
		strings.NewReader(string(loginBody)),
	)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var loginResp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&loginResp)

	if token, ok := loginResp["access_token"].(string); ok {
		return token, nil
	}

	return "", fmt.Errorf("login failed: no access token in response")
}

// CreateSessionRoom creates a new room for a session.
func (b *MatrixBroker) CreateSessionRoom(sessionID, name string) (string, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	createReq := map[string]interface{}{
		"name":   name,
		"topic":  fmt.Sprintf("Ayo session: %s", sessionID),
		"preset": "private_chat",
		"initial_state": []map[string]interface{}{
			{
				"type":      "m.room.history_visibility",
				"state_key": "",
				"content": map[string]string{
					"history_visibility": "shared",
				},
			},
		},
	}

	createBody, _ := json.Marshal(createReq)
	req, _ := http.NewRequest("POST", "http://localhost/_matrix/client/v3/createRoom", strings.NewReader(string(createBody)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+b.accessToken)

	resp, err := b.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("create room request: %w", err)
	}
	defer resp.Body.Close()

	var createResp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&createResp)

	roomID, ok := createResp["room_id"].(string)
	if !ok {
		return "", fmt.Errorf("create room failed: no room_id in response")
	}

	// Track room
	b.rooms[roomID] = &RoomInfo{
		ID:      roomID,
		Name:    name,
		Session: sessionID,
		Members: []string{},
	}

	return roomID, nil
}

// InviteAgent invites an agent to a room.
func (b *MatrixBroker) InviteAgent(roomID, handle string) error {
	// Ensure agent has Matrix user
	agent, err := b.EnsureAgentUser(handle)
	if err != nil {
		return err
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	// Invite
	inviteReq := map[string]string{
		"user_id": agent.UserID,
	}
	inviteBody, _ := json.Marshal(inviteReq)

	req, _ := http.NewRequest("POST",
		fmt.Sprintf("http://localhost/_matrix/client/v3/rooms/%s/invite", roomID),
		strings.NewReader(string(inviteBody)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+b.accessToken)

	resp, err := b.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("invite request: %w", err)
	}
	resp.Body.Close()

	// Auto-join on behalf of agent
	if err := b.joinRoomAsAgent(roomID, agent); err != nil {
		return fmt.Errorf("auto-join failed: %w", err)
	}

	// Update tracking
	if room, ok := b.rooms[roomID]; ok {
		room.Members = append(room.Members, handle)
	}

	return nil
}

func (b *MatrixBroker) joinRoomAsAgent(roomID string, agent *AgentUser) error {
	req, _ := http.NewRequest("POST",
		fmt.Sprintf("http://localhost/_matrix/client/v3/rooms/%s/join", roomID),
		strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+agent.Token)

	client := b.getAgentClient(agent.Handle)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func (b *MatrixBroker) getAgentClient(handle string) *http.Client {
	if client, ok := b.agentClients[handle]; ok {
		return client
	}

	transport := &http.Transport{
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			return net.Dial("unix", b.socketPath)
		},
	}
	client := &http.Client{Transport: transport, Timeout: 30 * time.Second}
	b.agentClients[handle] = client
	return client
}

// SendMessage sends a message to a room.
func (b *MatrixBroker) SendMessage(handle, roomID, content string) (string, error) {
	b.mu.RLock()
	agent, ok := b.agents[handle]
	b.mu.RUnlock()

	if !ok {
		return "", fmt.Errorf("agent not found: %s", handle)
	}

	msgReq := map[string]interface{}{
		"msgtype": "m.text",
		"body":    content,
	}
	msgBody, _ := json.Marshal(msgReq)

	txnID := fmt.Sprintf("%d", time.Now().UnixNano())
	req, _ := http.NewRequest("PUT",
		fmt.Sprintf("http://localhost/_matrix/client/v3/rooms/%s/send/m.room.message/%s", roomID, txnID),
		strings.NewReader(string(msgBody)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+agent.Token)

	client := b.getAgentClient(handle)
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	var sendResp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&sendResp)

	eventID, _ := sendResp["event_id"].(string)
	return eventID, nil
}

// SendMarkdown sends a markdown-formatted message.
func (b *MatrixBroker) SendMarkdown(handle, roomID, markdown string) (string, error) {
	b.mu.RLock()
	agent, ok := b.agents[handle]
	b.mu.RUnlock()

	if !ok {
		return "", fmt.Errorf("agent not found: %s", handle)
	}

	// Simple markdown to HTML conversion
	html := markdownToHTML(markdown)

	msgReq := map[string]interface{}{
		"msgtype":        "m.text",
		"body":           markdown,
		"format":         "org.matrix.custom.html",
		"formatted_body": html,
	}
	msgBody, _ := json.Marshal(msgReq)

	txnID := fmt.Sprintf("%d", time.Now().UnixNano())
	req, _ := http.NewRequest("PUT",
		fmt.Sprintf("http://localhost/_matrix/client/v3/rooms/%s/send/m.room.message/%s", roomID, txnID),
		strings.NewReader(string(msgBody)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+agent.Token)

	client := b.getAgentClient(handle)
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	var sendResp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&sendResp)

	eventID, _ := sendResp["event_id"].(string)
	return eventID, nil
}

// ListRooms returns all known rooms.
func (b *MatrixBroker) ListRooms() []*RoomInfo {
	b.mu.RLock()
	defer b.mu.RUnlock()

	rooms := make([]*RoomInfo, 0, len(b.rooms))
	for _, room := range b.rooms {
		rooms = append(rooms, room)
	}
	return rooms
}

// GetRoom returns a room by ID.
func (b *MatrixBroker) GetRoom(roomID string) (*RoomInfo, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	room, ok := b.rooms[roomID]
	if !ok {
		return nil, fmt.Errorf("room not found: %s", roomID)
	}
	return room, nil
}

// GetMessages retrieves queued messages for an agent.
func (b *MatrixBroker) GetMessages(handle string, limit int) []*QueuedMessage {
	b.mu.RLock()
	queue, ok := b.queues[handle]
	b.mu.RUnlock()

	if !ok {
		return nil
	}

	messages := make([]*QueuedMessage, 0, limit)
	for i := 0; i < limit; i++ {
		select {
		case msg := <-queue:
			messages = append(messages, msg)
		default:
			return messages
		}
	}
	return messages
}

// syncLoop runs the Matrix sync loop.
func (b *MatrixBroker) syncLoop(ctx context.Context) {
	b.mu.Lock()
	b.syncing = true
	b.mu.Unlock()

	defer func() {
		b.mu.Lock()
		b.syncing = false
		b.mu.Unlock()
	}()

	since := ""
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		// Build sync URL
		url := "http://localhost/_matrix/client/v3/sync?timeout=30000"
		if since != "" {
			url += "&since=" + since
		}

		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("Authorization", "Bearer "+b.accessToken)

		resp, err := b.httpClient.Do(req)
		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}

		var syncResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&syncResp)
		resp.Body.Close()

		// Process sync response
		if nextBatch, ok := syncResp["next_batch"].(string); ok {
			since = nextBatch
		}

		// Process room events
		if rooms, ok := syncResp["rooms"].(map[string]interface{}); ok {
			if join, ok := rooms["join"].(map[string]interface{}); ok {
				for roomID, roomData := range join {
					b.processRoomSync(roomID, roomData)
				}
			}
		}
	}
}

func (b *MatrixBroker) processRoomSync(roomID string, roomData interface{}) {
	data, ok := roomData.(map[string]interface{})
	if !ok {
		return
	}

	timeline, ok := data["timeline"].(map[string]interface{})
	if !ok {
		return
	}

	events, ok := timeline["events"].([]interface{})
	if !ok {
		return
	}

	for _, event := range events {
		evt, ok := event.(map[string]interface{})
		if !ok {
			continue
		}

		evtType, _ := evt["type"].(string)
		if evtType != "m.room.message" {
			continue
		}

		sender, _ := evt["sender"].(string)
		eventID, _ := evt["event_id"].(string)
		content, _ := evt["content"].(map[string]interface{})
		body, _ := content["body"].(string)
		ts, _ := evt["origin_server_ts"].(float64)

		// Ignore own messages
		if sender == b.daemonUserID {
			continue
		}

		msg := &QueuedMessage{
			RoomID:    roomID,
			EventID:   eventID,
			Sender:    sender,
			Content:   body,
			Timestamp: time.UnixMilli(int64(ts)),
		}

		// Extract handle from sender
		if strings.HasPrefix(sender, "@") && strings.Contains(sender, ":") {
			parts := strings.SplitN(sender[1:], ":", 2)
			msg.Handle = parts[0]
		}

		// Route message
		b.routeMessage(roomID, msg)
	}
}

func (b *MatrixBroker) routeMessage(roomID string, msg *QueuedMessage) {
	b.mu.RLock()
	room, ok := b.rooms[roomID]
	onMessage := b.onMessage
	b.mu.RUnlock()

	if !ok {
		return
	}

	// Route to all agents in room (except sender)
	for _, handle := range room.Members {
		if handle == msg.Handle {
			continue // Don't deliver own messages
		}

		// Check if message mentions this agent
		if b.shouldDeliverTo(msg, handle) {
			b.queueMessage(handle, msg)

			if onMessage != nil {
				go onMessage(handle, msg)
			}
		}
	}
}

func (b *MatrixBroker) shouldDeliverTo(msg *QueuedMessage, handle string) bool {
	// Always deliver if message mentions @handle
	if strings.Contains(msg.Content, "@"+handle) {
		return true
	}

	// Deliver if sender is not an agent (human message)
	agentDomain := ":" + b.serverName
	if !strings.HasSuffix(msg.Sender, agentDomain) {
		return true
	}

	return false
}

func (b *MatrixBroker) queueMessage(handle string, msg *QueuedMessage) {
	b.mu.Lock()
	defer b.mu.Unlock()

	queue, ok := b.queues[handle]
	if !ok {
		queue = make(chan *QueuedMessage, 100)
		b.queues[handle] = queue
	}

	select {
	case queue <- msg:
	default:
		// Queue full, drop oldest
		select {
		case <-queue:
		default:
		}
		queue <- msg
	}
}

// loadAgentCredentials loads saved agent credentials.
func (b *MatrixBroker) loadAgentCredentials() error {
	data, err := os.ReadFile(paths.MatrixAgentsFile())
	if err != nil {
		return err
	}

	var saved struct {
		Agents map[string]*AgentUser `json:"agents"`
	}
	if err := json.Unmarshal(data, &saved); err != nil {
		return err
	}

	b.agents = saved.Agents
	return nil
}

// saveAgentCredentials saves agent credentials.
func (b *MatrixBroker) saveAgentCredentials() error {
	saved := struct {
		Agents map[string]*AgentUser `json:"agents"`
	}{
		Agents: b.agents,
	}

	data, err := json.MarshalIndent(saved, "", "  ")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(paths.MatrixDataDir(), 0700); err != nil {
		return err
	}

	return os.WriteFile(paths.MatrixAgentsFile(), data, 0600)
}

// Helper functions

func generateRandomString(length int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = chars[time.Now().UnixNano()%int64(len(chars))]
		time.Sleep(time.Nanosecond)
	}
	return string(b)
}

func markdownToHTML(md string) string {
	// Very basic markdown to HTML conversion
	// In production, use a proper markdown library
	html := md
	
	// Code blocks
	html = strings.ReplaceAll(html, "```", "<pre><code>")
	
	// Bold
	for strings.Contains(html, "**") {
		html = strings.Replace(html, "**", "<strong>", 1)
		html = strings.Replace(html, "**", "</strong>", 1)
	}
	
	// Italic
	for strings.Contains(html, "*") {
		html = strings.Replace(html, "*", "<em>", 1)
		html = strings.Replace(html, "*", "</em>", 1)
	}
	
	return html
}
