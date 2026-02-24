package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/alexcabrera/ayo/internal/daemon"
)

func newNotificationsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "notifications",
		Aliases: []string{"notify", "notif"},
		Short:   "View trigger notifications",
		Long: `View notifications from completed trigger runs.

When ambient agents complete work in the background, notifications
are stored for later viewing.

Examples:
  # List unread notifications
  ayo notifications

  # List all notifications
  ayo notifications list --all

  # Mark all as read
  ayo notifications clear

  # Show notification details
  ayo notifications show 1`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return listNotificationsCmd().RunE(cmd, args)
		},
	}

	cmd.AddCommand(listNotificationsCmd())
	cmd.AddCommand(showNotificationCmd())
	cmd.AddCommand(clearNotificationsCmd())

	return cmd
}

func listNotificationsCmd() *cobra.Command {
	var all bool
	var limit int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List notifications",
		RunE: func(cmd *cobra.Command, args []string) error {
			ns, err := daemon.NewNotificationService("", daemon.DefaultNotificationConfig())
			if err != nil {
				return fmt.Errorf("open notifications: %w", err)
			}
			defer ns.Close()

			var notifications []*daemon.Notification
			if all {
				notifications, err = ns.GetAll(limit)
			} else {
				notifications, err = ns.GetUnread()
			}
			if err != nil {
				return fmt.Errorf("get notifications: %w", err)
			}

			if globalOutput.JSON {
				return json.NewEncoder(os.Stdout).Encode(notifications)
			}

			if len(notifications) == 0 {
				if !globalOutput.Quiet {
					if all {
						fmt.Println("No notifications")
					} else {
						fmt.Println("No unread notifications")
					}
				}
				return nil
			}

			// Styles
			purple := lipgloss.Color("#a78bfa")
			muted := lipgloss.Color("#6b7280")
			subtle := lipgloss.Color("#374151")
			green := lipgloss.Color("#34d399")
			red := lipgloss.Color("#f87171")

			headerStyle := lipgloss.NewStyle().Bold(true).Foreground(purple)
			dividerStyle := lipgloss.NewStyle().Foreground(subtle)
			mutedStyle := lipgloss.NewStyle().Foreground(muted)
			successStyle := lipgloss.NewStyle().Foreground(green)
			failedStyle := lipgloss.NewStyle().Foreground(red)

			title := "Unread Notifications"
			if all {
				title = "All Notifications"
			}

			fmt.Println()
			fmt.Println(headerStyle.Render("  📬 " + title))
			fmt.Println(dividerStyle.Render("  " + strings.Repeat("─", 60)))
			fmt.Println()

			for _, n := range notifications {
				statusIcon := successStyle.Render("✓")
				if n.Status == daemon.NotificationStatusFailed {
					statusIcon = failedStyle.Render("✗")
				}

				timeAgo := formatTimeAgo(n.CreatedAt.Unix())
				readMark := ""
				if n.ReadAt != nil {
					readMark = mutedStyle.Render(" (read)")
				}

				fmt.Printf("  %d. %s %s %s%s\n", n.ID, statusIcon, n.TriggerName, mutedStyle.Render(timeAgo), readMark)
				if n.Summary != "" {
					summary := n.Summary
					if len(summary) > 55 {
						summary = summary[:52] + "..."
					}
					fmt.Printf("     %s\n", mutedStyle.Render(summary))
				}
				fmt.Println()
			}

			fmt.Println(dividerStyle.Render("  " + strings.Repeat("─", 60)))
			unreadCount, _ := ns.UnreadCount()
			if unreadCount > 0 {
				fmt.Println(mutedStyle.Render(fmt.Sprintf("  %d unread • Run 'ayo notifications clear' to mark as read", unreadCount)))
			} else {
				fmt.Println(mutedStyle.Render(fmt.Sprintf("  %d notifications", len(notifications))))
			}
			fmt.Println()

			return nil
		},
	}

	cmd.Flags().BoolVarP(&all, "all", "a", false, "show all notifications (including read)")
	cmd.Flags().IntVarP(&limit, "limit", "n", 50, "maximum notifications to show")

	return cmd
}

func showNotificationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show notification details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var id int64
			if _, err := fmt.Sscanf(args[0], "%d", &id); err != nil {
				return fmt.Errorf("invalid notification ID: %s", args[0])
			}

			ns, err := daemon.NewNotificationService("", daemon.DefaultNotificationConfig())
			if err != nil {
				return fmt.Errorf("open notifications: %w", err)
			}
			defer ns.Close()

			n, err := ns.Get(id)
			if err != nil {
				return err
			}

			// Mark as read when viewed
			ns.MarkRead(id)

			if globalOutput.JSON {
				return json.NewEncoder(os.Stdout).Encode(n)
			}

			// Styles
			purple := lipgloss.Color("#a78bfa")
			cyan := lipgloss.Color("#67e8f9")
			muted := lipgloss.Color("#6b7280")
			subtle := lipgloss.Color("#374151")
			green := lipgloss.Color("#34d399")
			red := lipgloss.Color("#f87171")
			text := lipgloss.Color("#e5e7eb")

			headerStyle := lipgloss.NewStyle().Bold(true).Foreground(purple)
			dividerStyle := lipgloss.NewStyle().Foreground(subtle)
			iconStyle := lipgloss.NewStyle().Foreground(cyan)
			labelStyle := lipgloss.NewStyle().Foreground(muted)
			valueStyle := lipgloss.NewStyle().Foreground(text)

			statusIcon := green
			statusText := "Success"
			if n.Status == daemon.NotificationStatusFailed {
				statusIcon = red
				statusText = "Failed"
			}

			fmt.Println()
			fmt.Println("  " + iconStyle.Render("📬") + " " + headerStyle.Render(n.TriggerName))
			fmt.Println(dividerStyle.Render("  " + strings.Repeat("─", 50)))

			fmt.Printf("  %s   %s\n", labelStyle.Render("Status:"), lipgloss.NewStyle().Foreground(statusIcon).Render(statusText))
			fmt.Printf("  %s     %s\n", labelStyle.Render("Time:"), valueStyle.Render(n.CreatedAt.Format("2006-01-02 15:04:05")))

			if n.Summary != "" {
				fmt.Printf("  %s  %s\n", labelStyle.Render("Summary:"), valueStyle.Render(n.Summary))
			}

			if n.OutputPath != "" {
				fmt.Printf("  %s   %s\n", labelStyle.Render("Output:"), valueStyle.Render(n.OutputPath))
			}

			fmt.Println()

			return nil
		},
	}

	return cmd
}

func clearNotificationsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clear",
		Short: "Mark all notifications as read",
		RunE: func(cmd *cobra.Command, args []string) error {
			ns, err := daemon.NewNotificationService("", daemon.DefaultNotificationConfig())
			if err != nil {
				return fmt.Errorf("open notifications: %w", err)
			}
			defer ns.Close()

			if err := ns.MarkAllRead(); err != nil {
				return fmt.Errorf("mark as read: %w", err)
			}

			if !globalOutput.Quiet {
				successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
				fmt.Println(successStyle.Render("✓ All notifications marked as read"))
			}

			return nil
		},
	}

	return cmd
}
