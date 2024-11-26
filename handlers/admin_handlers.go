package handlers

import (
	"encoding/json"
	"net/http"

	"gorm.io/gorm"

	"github.com/mbsoft31/screens/models"
)

type Ctx string

const CtxDb Ctx = "db"

// Helper function to retrieve the database from context
func getDB(r *http.Request) *gorm.DB {
	return r.Context().Value(CtxDb).(*gorm.DB)
}

// ScreensHandler handles GET and POST requests for screens
func ScreensHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		GetScreens(w, r)
	case http.MethodPost:
		CreateScreen(w, r)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

// PlaylistsHandler handles GET and POST requests for playlists
func PlaylistsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		GetPlaylists(w, r)
	case http.MethodPost:
		CreatePlaylist(w, r)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

// UsersHandler handles GET and POST requests for users
func UsersHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		GetUsers(w, r)
	case http.MethodPost:
		CreateUser(w, r)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

// GetScreens fetches all screens
func GetScreens(w http.ResponseWriter, r *http.Request) {
	db := getDB(r)

	var screens []models.Screen
	if err := db.Preload("Location").Preload("Playlists.Items").Find(&screens).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(screens)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// CreateScreen creates a new screen
func CreateScreen(w http.ResponseWriter, r *http.Request) {
	db := getDB(r)

	var screen models.Screen
	if err := json.NewDecoder(r.Body).Decode(&screen); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	if err := db.Create(&screen).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	err := json.NewEncoder(w).Encode(screen)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// GetPlaylists fetches all playlists
func GetPlaylists(w http.ResponseWriter, r *http.Request) {
	db := getDB(r)

	var playlists []models.Playlist
	if err := db.Preload("Items").Find(&playlists).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(playlists)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// CreatePlaylist creates a new playlist
func CreatePlaylist(w http.ResponseWriter, r *http.Request) {
	db := getDB(r)

	var playlist models.Playlist
	if err := json.NewDecoder(r.Body).Decode(&playlist); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	if err := db.Create(&playlist).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	err := json.NewEncoder(w).Encode(playlist)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// GetUsers fetches all users
func GetUsers(w http.ResponseWriter, r *http.Request) {
	db := getDB(r)

	var users []models.User
	if err := db.Preload("Roles.Permissions").Preload("Permissions").Find(&users).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(users)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// CreateUser creates a new user
func CreateUser(w http.ResponseWriter, r *http.Request) {
	db := getDB(r)

	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	if err := db.Create(&user).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	err := json.NewEncoder(w).Encode(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
