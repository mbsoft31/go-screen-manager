package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/mbsoft31/screens/models"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
)

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:8081")
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer func(conn net.Conn) {
		_ = conn.Close()
	}(conn)

	fmt.Println("Connected to server")

	clientID := "1"
	localPlaylistHash := ""

	for {
		_, err := fmt.Fprintln(conn, clientID)
		if err != nil {
			log.Println("Failed to send client ID:", err)
			break
		}

		fmt.Println("Client ID: ", clientID)

		// Read playlist data from server
		data, err := io.ReadAll(conn)
		if err != nil {
			fmt.Println("Error reading data from server:", err)
			break
		}

		fmt.Printf("Received data: %s\n", string(data))

		var playlist models.Playlist
		if err := json.Unmarshal(data, &playlist); err != nil {
			fmt.Println("Error unmarshalling playlist:", err)
			continue
		}

		// Calculate hash of the received playlist
		newHash := calculateHash(playlist)
		if newHash == localPlaylistHash {
			fmt.Println("Playlist unchanged")
			continue
		}

		fmt.Println("Playlist updated")
		localPlaylistHash = newHash

		// Check for new items
		/*newItems := getNewItems(playlist.Items)
		for _, item := range newItems {
			downloadMedia(item)
		}*/

		// Play the updated playlist
		// playPlaylist(playlist)
	}
}

func calculateHash(playlist models.Playlist) string {
	data, _ := json.Marshal(playlist)
	return fmt.Sprintf("%x", md5.Sum(data))
}

func getNewItems(serverItems []models.Media) []string {
	localFiles := listLocalMedia()
	var newItems []string
	for _, item := range serverItems {
		if !contains(localFiles, item.Path) {
			newItems = append(newItems, item.Path)
		}
	}
	return newItems
}

func listLocalMedia() []string {
	files, _ := os.ReadDir("./media")
	var fileList []string
	for _, file := range files {
		fileList = append(fileList, file.Name())
	}
	return fileList
}

func contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

func downloadMedia(url string) {
	fmt.Println("Downloading media:", url)
	cmd := exec.Command("curl", "-O", url) // Adjust download logic if needed
	cmd.Dir = "./media"
	_ = cmd.Run()
}

func playPlaylist(playlist models.Playlist) {
	fmt.Println("Playing playlist...")
	// Stop VLC if running
	err := exec.Command("pkill", "vlc").Run()
	if err != nil {
		return
	}

	// Start VLC with the new playlist
	args := []string{"--fullscreen"}
	for _, item := range playlist.Items {
		args = append(args, "./media/"+item.Path) // Assuming files are stored in ./media
	}
	cmd := exec.Command("vlc", args...)
	_ = cmd.Start()
}
