package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const (
	url           = "https://jlistopad.cz"
	checkInterval = 10 * time.Minute
	cacheFile     = "last_content.txt"
	ntfyTopic     = "jlistopad"
)

// Fetches the webpage and extracts the content inside the .panel-body div
func fetchPageContent() (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("error: received status code %d", resp.StatusCode)
	}

	// Load HTML into goquery
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", err
	}

	// Extract content inside the .panel-body div
	content := doc.Find(".panel-body").Text()
	return content, nil
}

// Sends a push notification via ntfy.sh
func sendPushNotification(message string) {
	ntfyURL := "https://ntfy.sh/" + ntfyTopic

	req, err := http.NewRequest("POST", ntfyURL, bytes.NewBufferString(message))
	if err != nil {
		log.Println("Error creating request:", err)
		return
	}
	req.Header.Set("Content-Type", "text/plain")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Failed to send notification:", err)
		return
	}
	defer resp.Body.Close()

	log.Println("Push notification sent via ntfy.sh!")
}

// Reads the last cached content from a file
func readCache() string {
	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return ""
	}
	return string(data)
}

// Writes new content to the cache file
func writeCache(content string) {
	_ = os.WriteFile(cacheFile, []byte(content), 0644)
}

func main() {
	log.Println("Starting website monitor...")

	// Send a startup notification
	sendPushNotification("🔄 Website monitor service restarted!")

	for {
		content, err := fetchPageContent()
		if err != nil {
			log.Println("Error fetching page content:", err)
			time.Sleep(checkInterval)
			continue
		}

		lastContent := readCache()

		if lastContent != "" && content != lastContent {
			log.Println("Website content changed! Sending notification...")
			sendPushNotification("📢 The website has been updated: " + url)
		} else {
			log.Println("No change detected.")
		}

		writeCache(content)
		time.Sleep(checkInterval)
	}
}