package client_manager

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/mbsoft31/screens/db"
	"github.com/mbsoft31/screens/models"
	"io"
	"net"
	"strings"
	"sync"
	"time"
)

//TODO: fix the client is stuck before sending playlist after i added hash versioning

type PlaylistVersion struct {
	VersionHash string
	Playlist    models.Playlist
}

type ClientManager struct {
	Mutex sync.Mutex

	Listener         net.Listener
	Connections      []net.Conn
	PlaylistVersions map[string]PlaylistVersion
}

func NewManager(address string) (*ClientManager, error) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("error starting TCP server: %w", err)
	}

	fmt.Printf("TCP Server is running on %s\n", address)
	return &ClientManager{
		Listener:         listener,
		Connections:      make([]net.Conn, 0),
		PlaylistVersions: make(map[string]PlaylistVersion),
	}, nil
}

func (cm *ClientManager) Listen() {
	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			fmt.Println("Error closing TCP server:", err)
		}
	}(cm.Listener)

	for {
		conn, err := cm.Listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		// Enable TCP keep-alive
		tcpConn, ok := conn.(*net.TCPConn)
		if ok {
			_ = tcpConn.SetKeepAlive(true)
			_ = tcpConn.SetKeepAlivePeriod(30 * time.Second)
		}

		cm.AddConnection(conn)
		go cm.HandleClient(conn)
	}
}

func (cm *ClientManager) AddConnection(conn net.Conn) net.Conn {
	cm.Mutex.Lock()
	defer cm.Mutex.Unlock()
	cm.Connections = append(cm.Connections, conn)
	return conn
}

func (cm *ClientManager) RemoveConnection(conn net.Conn) {
	cm.Mutex.Lock()
	defer cm.Mutex.Unlock()
	for i, c := range cm.Connections {
		if c == conn {
			cm.Connections = append(cm.Connections[:i], cm.Connections[i+1:]...)
			return
		}
	}
}

func (cm *ClientManager) RemoveAllConnections() {
	cm.Mutex.Lock()
	defer cm.Mutex.Unlock()
	for _, conn := range cm.Connections {
		_ = conn.Close() // Close each connection
	}
	cm.Connections = make([]net.Conn, 0)
}

func (cm *ClientManager) HandleClient(conn net.Conn) {
	fmt.Printf("New connection from %s\n", conn.RemoteAddr().String())

	defer func() {
		cm.RemoveConnection(conn)
		_ = conn.Close()
		fmt.Println("Connection closed:", conn.RemoteAddr())
	}()

	// Read client ID from the connection
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		if err == io.EOF {
			fmt.Println("Client disconnected:", conn.RemoteAddr())
		} else {
			fmt.Println("Error reading from client:", err)
		}
		return
	}

	clientID := strings.TrimSpace(string(buf[:n]))
	if !isValidClientID(clientID) {
		fmt.Println("Invalid client ID:", clientID)
		return
	}

	fmt.Printf("Client %s connected\n", clientID)

	// Initialize the client's playlist hash
	{
		cm.Mutex.Lock()
		defer cm.Mutex.Unlock()
		if _, exists := cm.PlaylistVersions[clientID]; !exists {
			fmt.Println("Playlist version not found:", clientID)
			cm.PlaylistVersions[clientID] = PlaylistVersion{}
		}
	}

	// Handle playlist updates for this client
	for {
		fmt.Printf("handle for %s", clientID)
		go cm.HandlePlaylistUpdate(clientID, conn)
		time.Sleep(5 * time.Second) // Poll for updates every 5 seconds
	}
}

func isValidClientID(id string) bool {
	return len(id) > 0 && len(id) <= 36 // Adjust the rules as needed
}

func (cm *ClientManager) HandlePlaylistUpdate(clientID string, conn net.Conn) {
	/*cm.Mutex.Lock()
	defer cm.Mutex.Unlock()*/
	fmt.Printf("tttttttttttttt")

	// Fetch the current playlist for the client
	var currentPlaylist models.Playlist
	if err := db.DB.
		Preload("Items").
		Where("screen_id = ?", clientID).
		First(&currentPlaylist).Error; err != nil {
		fmt.Println("Error fetching playlist for client ID:", clientID, err)
		return
	}

	fmt.Printf("Found playlist for client %s, {%v}\n", clientID, currentPlaylist)

	// Calculate a hash of the current playlist
	newVersionHash := calculateHash(currentPlaylist)

	fmt.Printf("New version hash: %s\n", newVersionHash)
	fmt.Printf("Version hash: %s\n", cm.PlaylistVersions[clientID].VersionHash)
	// Check if the playlist has changed
	if cm.PlaylistVersions[clientID].VersionHash == newVersionHash {
		fmt.Println("Skipped update")
		return // No changes, skip update
	}

	// Update the stored version and send the updated playlist
	cm.PlaylistVersions[clientID] = PlaylistVersion{
		VersionHash: newVersionHash,
		Playlist:    currentPlaylist,
	}

	data, err := json.Marshal(currentPlaylist)
	if err != nil {
		fmt.Println("Error marshalling playlist for client ID:", clientID, err)
		return
	}

	response := append(data, '\n')
	fmt.Printf("Sending data to client: %s", response)
	if _, err := conn.Write(response); err != nil {
		fmt.Println("Error sending updated playlist to client:", err)
	}
}

func calculateHash(playlist models.Playlist) string {
	data, _ := json.Marshal(playlist)
	return fmt.Sprintf("%x", md5.Sum(data)) // Use any hash method
}
