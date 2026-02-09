package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/alexcabrera/ayo/internal/daemon"
	"github.com/spf13/cobra"
)

// newMatrixCmd creates the matrix command for inter-agent communication.
func newMatrixCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "matrix",
		Aliases: []string{"m"},
		Short:   "Matrix chat commands for inter-agent communication",
		Long: `Communicate with other agents via Matrix.

These commands connect through the daemon to the local Matrix homeserver (Conduit).
They work both on the host and inside sandboxes.`,
	}

	cmd.AddCommand(newMatrixStatusCmd())
	cmd.AddCommand(newMatrixRoomsCmd())
	cmd.AddCommand(newMatrixCreateCmd())
	cmd.AddCommand(newMatrixSendCmd())
	cmd.AddCommand(newMatrixReadCmd())
	cmd.AddCommand(newMatrixWhoCmd())
	cmd.AddCommand(newMatrixInviteCmd())

	return cmd
}

func newMatrixStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show Matrix connection status",
		RunE:  runMatrixStatus,
	}
}

func newMatrixRoomsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rooms",
		Short: "List Matrix rooms",
		RunE:  runMatrixRooms,
	}
	cmd.Flags().String("session", "", "Filter by session ID")
	return cmd
}

func newMatrixCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a Matrix room",
		RunE:  runMatrixCreate,
	}
	cmd.Flags().StringP("name", "n", "", "Room name (required)")
	cmd.Flags().String("session", "", "Associate with session ID")
	cmd.Flags().StringSlice("invite", nil, "Agents to invite")
	cmd.MarkFlagRequired("name")
	return cmd
}

func newMatrixSendCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send <room> <message>",
		Short: "Send a message to a room",
		Args:  cobra.MinimumNArgs(1),
		RunE:  runMatrixSend,
	}
	cmd.Flags().StringP("file", "f", "", "Read message from file")
	cmd.Flags().String("as", "", "Send as specific agent")
	cmd.Flags().Bool("markdown", true, "Parse as markdown")
	return cmd
}

func newMatrixReadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "read <room>",
		Short: "Read messages from a room",
		Args:  cobra.ExactArgs(1),
		RunE:  runMatrixRead,
	}
	cmd.Flags().IntP("limit", "n", 20, "Number of messages")
	cmd.Flags().BoolP("follow", "f", false, "Stream new messages")
	return cmd
}

func newMatrixWhoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "who <room>",
		Short: "List room members",
		Args:  cobra.ExactArgs(1),
		RunE:  runMatrixWho,
	}
}

func newMatrixInviteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "invite <room> <agent>",
		Short: "Invite an agent to a room",
		Args:  cobra.ExactArgs(2),
		RunE:  runMatrixInvite,
	}
}

func connectDaemon(ctx context.Context) (*daemon.Client, error) {
	client := daemon.NewClient()
	if err := client.Connect(ctx); err != nil {
		return nil, fmt.Errorf("connect to daemon: %w", err)
	}
	return client, nil
}

func runMatrixStatus(cmd *cobra.Command, args []string) error {
	client, err := connectDaemon(cmd.Context())
	if err != nil {
		return err
	}
	defer client.Close()

	result, err := client.MatrixStatus(cmd.Context())
	if err != nil {
		return err
	}

	if globalOutput.JSON {
		return json.NewEncoder(os.Stdout).Encode(result)
	}

	fmt.Println("Matrix Status:")
	fmt.Println()

	// Conduit status
	fmt.Println("Conduit:")
	if result.Conduit.Running {
		fmt.Printf("  Running: yes (PID %d)\n", result.Conduit.Pid)
		fmt.Printf("  Healthy: %v\n", result.Conduit.Healthy)
		fmt.Printf("  Uptime: %ds\n", result.Conduit.Uptime)
		fmt.Printf("  Restarts: %d\n", result.Conduit.Restarts)
	} else {
		fmt.Println("  Running: no")
	}
	fmt.Println()

	// Broker status
	fmt.Println("Broker:")
	fmt.Printf("  Connected: %v\n", result.Broker.Connected)
	fmt.Printf("  Syncing: %v\n", result.Broker.Syncing)
	fmt.Printf("  Rooms: %d\n", result.Broker.RoomCount)
	fmt.Printf("  Agents: %d\n", result.Broker.AgentCount)
	fmt.Printf("  Queued: %d\n", result.Broker.QueuedMsgs)

	return nil
}

func runMatrixRooms(cmd *cobra.Command, args []string) error {
	client, err := connectDaemon(cmd.Context())
	if err != nil {
		return err
	}
	defer client.Close()

	sessionID, _ := cmd.Flags().GetString("session")

	params := daemon.MatrixRoomsListParams{
		SessionID: sessionID,
	}

	result, err := client.MatrixRoomsList(cmd.Context(), params)
	if err != nil {
		return err
	}

	if globalOutput.JSON {
		return json.NewEncoder(os.Stdout).Encode(result)
	}

	if len(result.Rooms) == 0 {
		fmt.Println("No rooms found")
		return nil
	}

	// Table header
	fmt.Printf("%-30s %-15s %-8s %s\n", "ROOM", "SESSION", "MEMBERS", "UNREAD")
	fmt.Println(strings.Repeat("-", 70))

	for _, room := range result.Rooms {
		session := room.Session
		if session == "" {
			session = "-"
		}
		fmt.Printf("%-30s %-15s %-8d %d\n", room.Name, session, len(room.Members), room.Unread)
	}

	return nil
}

func runMatrixCreate(cmd *cobra.Command, args []string) error {
	client, err := connectDaemon(cmd.Context())
	if err != nil {
		return err
	}
	defer client.Close()

	name, _ := cmd.Flags().GetString("name")
	sessionID, _ := cmd.Flags().GetString("session")
	invite, _ := cmd.Flags().GetStringSlice("invite")

	params := daemon.MatrixRoomsCreateParams{
		Name:      name,
		SessionID: sessionID,
		Invite:    invite,
	}

	result, err := client.MatrixRoomsCreate(cmd.Context(), params)
	if err != nil {
		return err
	}

	if globalOutput.JSON {
		return json.NewEncoder(os.Stdout).Encode(result)
	}

	fmt.Printf("Created room %s (%s)\n", name, result.RoomID)
	return nil
}

func runMatrixSend(cmd *cobra.Command, args []string) error {
	client, err := connectDaemon(cmd.Context())
	if err != nil {
		return err
	}
	defer client.Close()

	roomID := args[0]
	var content string

	// Get message content
	filePath, _ := cmd.Flags().GetString("file")
	if filePath != "" {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("read file: %w", err)
		}
		content = string(data)
	} else if len(args) > 1 && args[1] == "-" {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("read stdin: %w", err)
		}
		content = string(data)
	} else if len(args) > 1 {
		content = strings.Join(args[1:], " ")
	} else {
		return fmt.Errorf("no message provided")
	}

	asAgent, _ := cmd.Flags().GetString("as")
	if asAgent == "" {
		asAgent = os.Getenv("AYO_AGENT_HANDLE")
	}

	markdown, _ := cmd.Flags().GetBool("markdown")
	format := "plain"
	if markdown {
		format = "markdown"
	}

	params := daemon.MatrixSendParams{
		RoomID:  roomID,
		Content: content,
		AsAgent: asAgent,
		Format:  format,
	}

	result, err := client.MatrixSend(cmd.Context(), params)
	if err != nil {
		return err
	}

	if globalOutput.JSON {
		return json.NewEncoder(os.Stdout).Encode(result)
	}

	if !globalOutput.Quiet {
		fmt.Printf("Sent message (%s)\n", result.EventID)
	}
	return nil
}

func runMatrixRead(cmd *cobra.Command, args []string) error {
	client, err := connectDaemon(cmd.Context())
	if err != nil {
		return err
	}
	defer client.Close()

	roomID := args[0]
	limit, _ := cmd.Flags().GetInt("limit")
	follow, _ := cmd.Flags().GetBool("follow")

	params := daemon.MatrixReadParams{
		RoomID: roomID,
		Limit:  limit,
	}

	if follow {
		// Streaming mode
		return runMatrixReadFollow(cmd.Context(), client, roomID)
	}

	result, err := client.MatrixRead(cmd.Context(), params)
	if err != nil {
		return err
	}

	if globalOutput.JSON {
		return json.NewEncoder(os.Stdout).Encode(result)
	}

	if len(result.Messages) == 0 {
		fmt.Println("No messages")
		return nil
	}

	for _, msg := range result.Messages {
		printMatrixMessage(msg)
	}

	return nil
}

func runMatrixReadFollow(ctx context.Context, client *daemon.Client, roomID string) error {
	since := ""
	for {
		params := daemon.MatrixReadParams{
			RoomID: roomID,
			Limit:  10,
			After:  since,
		}

		result, err := client.MatrixRead(ctx, params)
		if err != nil {
			return err
		}

		for _, msg := range result.Messages {
			if globalOutput.JSON {
				json.NewEncoder(os.Stdout).Encode(msg)
			} else {
				printMatrixMessage(msg)
			}
			since = msg.EventID
		}

		time.Sleep(2 * time.Second)
	}
}

func printMatrixMessage(msg *daemon.QueuedMessage) {
	ts := msg.Timestamp.Format("15:04:05")
	sender := msg.Handle
	if sender == "" {
		sender = msg.Sender
	}
	fmt.Printf("[%s] @%s: %s\n", ts, sender, msg.Content)
}

func runMatrixWho(cmd *cobra.Command, args []string) error {
	client, err := connectDaemon(cmd.Context())
	if err != nil {
		return err
	}
	defer client.Close()

	roomID := args[0]

	params := daemon.MatrixRoomsMembersParams{
		RoomID: roomID,
	}

	result, err := client.MatrixRoomsMembers(cmd.Context(), params)
	if err != nil {
		return err
	}

	if globalOutput.JSON {
		return json.NewEncoder(os.Stdout).Encode(result)
	}

	if len(result.Members) == 0 {
		fmt.Println("No members")
		return nil
	}

	fmt.Printf("%-20s %-10s\n", "MEMBER", "TYPE")
	fmt.Println(strings.Repeat("-", 35))

	for _, member := range result.Members {
		memberType := "user"
		if member.IsAgent {
			memberType = "agent"
		}
		fmt.Printf("%-20s %-10s\n", member.DisplayName, memberType)
	}

	return nil
}

func runMatrixInvite(cmd *cobra.Command, args []string) error {
	client, err := connectDaemon(cmd.Context())
	if err != nil {
		return err
	}
	defer client.Close()

	roomID := args[0]
	handle := strings.TrimPrefix(args[1], "@")

	params := daemon.MatrixRoomsInviteParams{
		RoomID: roomID,
		Handle: handle,
	}

	if err := client.MatrixRoomsInvite(cmd.Context(), params); err != nil {
		return err
	}

	if !globalOutput.Quiet {
		fmt.Printf("Invited @%s to %s\n", handle, roomID)
	}
	return nil
}
