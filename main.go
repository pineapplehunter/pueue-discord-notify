package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
)

type WebhookPayload struct {
	Content string  `json:"content"`
	Embeds  []Embed `json:"embeds,omitempty"`
}

type Embed struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Color       int     `json:"color"`
	Fields      []Field `json:"fields"`
}

type Field struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

func resultColor(result string) int {
	switch result {
	case "Success":
		return 0x57F287
	case "Failed", "Errored":
		return 0xED4245
	case "Killed":
		return 0xFEE75C
	default:
		return 0x5865F2
	}
}

func sendNotification(webhookURL, taskID, command, result, exitCode, group, host string) error {
	desc := command
	if len(desc) > 2000 {
		desc = desc[:2000]
	}

	fields := []Field{
		{Name: "Result", Value: result, Inline: true},
		{Name: "Exit Code", Value: exitCode, Inline: true},
	}
	if group != "" {
		fields = append(fields, Field{Name: "Group", Value: group, Inline: true})
	}

	fields = append([]Field{
		{Name: "Host", Value: host, Inline: true},
	}, fields...)

	payload := WebhookPayload{
		Content: fmt.Sprintf("[%s] Task #%s **%s**", host, taskID, result),
		Embeds: []Embed{
			{
				Title:       fmt.Sprintf("[%s] Task #%s", host, taskID),
				Description: "```\n" + desc + "\n```",
				Color:       resultColor(result),
				Fields:      fields,
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshaling payload: %w", err)
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("sending webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned %s", resp.Status)
	}

	return nil
}

func run() error {
	webhookFile := flag.String("webhook-file", "", "Path to file containing Discord webhook URL")
	taskID := flag.String("id", "", "Pueue task ID")
	command := flag.String("command", "", "Pueue task command")
	result := flag.String("result", "", "Pueue task result")
	exitCode := flag.String("exit-code", "", "Pueue task exit code")
	group := flag.String("group", "", "Pueue task group")
	host := flag.String("host", "", "Host label (defaults to system hostname)")
	flag.Parse()

	if *host == "" {
		h, err := os.Hostname()
		if err != nil {
			return fmt.Errorf("getting hostname: %w", err)
		}
		*host = h
	}

	if *webhookFile == "" {
		return fmt.Errorf("--webhook-file is required")
	}

	data, err := os.ReadFile(*webhookFile)
	if err != nil {
		return fmt.Errorf("reading webhook file: %w", err)
	}
	webhookURL := string(bytes.TrimSpace(data))
	if webhookURL == "" {
		return fmt.Errorf("webhook file is empty")
	}

	return sendNotification(webhookURL, *taskID, *command, *result, *exitCode, *group, *host)
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
