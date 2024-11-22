package models

import (
	"encoding/json"
	"errors"
	"fmt"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// User model represents system users.
type User struct {
	gorm.Model
	Username    string       `gorm:"uniqueIndex;size:50" validate:"required"`
	Email       string       `gorm:"uniqueIndex" validate:"required,email"`
	Password    string       `gorm:"size:100" validate:"required,min=8"` // Store hashed password
	Roles       []Role       `gorm:"many2many:user_roles"`
	Permissions []Permission `gorm:"many2many:user_permissions"`
}

// Role model represents user roles.
type Role struct {
	gorm.Model
	Name        string       `gorm:"uniqueIndex;size:50" validate:"required"`
	Guard       string       `gorm:"default:web;size:20"`
	Permissions []Permission `gorm:"many2many:role_permissions"`
}

// Permission model represents system permissions.
type Permission struct {
	gorm.Model
	Name  string `gorm:"uniqueIndex;size:50" validate:"required"`
	Guard string `gorm:"default:web;size:20"`
}

// Media model represents media files.
type Media struct {
	gorm.Model
	Name       string `gorm:"size:100"`
	Path       string `gorm:"size:255" validate:"required"`
	Type       string `gorm:"size:50"`
	PlaylistID uint   `gorm:"not null"` // Foreign key for Playlist
}

// Playlist model represents playlists containing media items.
type Playlist struct {
	gorm.Model
	Name     string  `gorm:"size:100" validate:"required"`
	Items    []Media `gorm:"constraint:OnDelete:CASCADE"`
	ScreenID uint    `gorm:"not null"` // Foreign key for Screen
}

// Location model represents a geographic location.
type Location struct {
	gorm.Model
	ScreenID uint    `gorm:"not null"` // Foreign key for Screen
	Long     float64 `gorm:"not null"`
	Lat      float64 `gorm:"not null"`
}

// Screen model represents a display screen.
type Screen struct {
	gorm.Model
	Name      string     `gorm:"size:100" validate:"required"`
	Location  Location   `gorm:"constraint:OnDelete:CASCADE"`
	Playlists []Playlist `gorm:"constraint:OnDelete:CASCADE"`
	Meta      datatypes.JSON
}

// Helper Methods for Screen Meta Field

// GetMeta parses the Meta JSON field into a map.
func (s *Screen) GetMeta() (map[string]interface{}, error) {
	var meta map[string]interface{}
	if err := json.Unmarshal(s.Meta, &meta); err != nil {
		return nil, err
	}
	return meta, nil
}

// SetMeta sets the Meta JSON field from a map.
func (s *Screen) SetMeta(meta map[string]interface{}) error {
	data, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	s.Meta = data
	return nil
}

// Validation Hook Examples

// BeforeCreate hook to ensure default values for Role and Permission.
func (r *Role) BeforeCreate(*gorm.DB) (err error) {
	if r.Guard == "" {
		r.Guard = "web"
	}
	return
}

func (p *Permission) BeforeCreate(*gorm.DB) (err error) {
	if p.Guard == "" {
		p.Guard = "web"
	}
	return
}

// BeforeSave Validation Hook for User to ensure non-empty password.
func (u *User) BeforeSave(*gorm.DB) (err error) {
	if u.Password == "" {
		return errors.New("password cannot be empty")
	}
	return
}

// Seed function populates the database with test data.
func Seed(db *gorm.DB) error {
	// Delete existing data to start fresh
	if err := db.Migrator().DropTable(
		&User{}, &Role{}, &Permission{},
		&Media{}, &Playlist{}, &Location{}, &Screen{},
	); err != nil {
		return fmt.Errorf("failed to drop tables: %w", err)
	}

	// Auto-migrate models
	if err := db.AutoMigrate(
		&User{}, &Role{}, &Permission{},
		&Media{}, &Playlist{}, &Location{}, &Screen{},
	); err != nil {
		return fmt.Errorf("failed to migrate models: %w", err)
	}

	// Seed Permissions
	permissions := []Permission{
		{Name: "view_dashboard", Guard: "web"},
		{Name: "edit_users", Guard: "web"},
		{Name: "delete_posts", Guard: "web"},
	}
	if err := db.Create(&permissions).Error; err != nil {
		return fmt.Errorf("failed to seed permissions: %w", err)
	}

	// Seed Roles
	roles := []Role{
		{Name: "admin", Guard: "web", Permissions: permissions},
		{Name: "editor", Guard: "web", Permissions: permissions[:2]},
	}
	if err := db.Create(&roles).Error; err != nil {
		return fmt.Errorf("failed to seed roles: %w", err)
	}

	// Seed Users
	users := []User{
		{Username: "admin", Email: "admin@example.com", Password: "hashed_password", Roles: roles[:1], Permissions: permissions},
		{Username: "editor", Email: "editor@example.com", Password: "hashed_password", Roles: roles[1:], Permissions: permissions[:2]},
	}
	if err := db.Create(&users).Error; err != nil {
		return fmt.Errorf("failed to seed users: %w", err)
	}

	// Seed Media
	mediaItems := []Media{
		{Name: "Image1", Path: "/media/image1.png", Type: "image"},
		{Name: "Video1", Path: "/media/video1.mp4", Type: "video"},
	}
	if err := db.Create(&mediaItems).Error; err != nil {
		return fmt.Errorf("failed to seed media: %w", err)
	}

	// Seed Playlists
	playlists := []Playlist{
		{Name: "Playlist1", Items: mediaItems},
	}
	if err := db.Create(&playlists).Error; err != nil {
		return fmt.Errorf("failed to seed playlists: %w", err)
	}

	// Seed Screens
	screens := []Screen{
		{
			Name:      "Screen1",
			Location:  Location{Long: 36.7783, Lat: -119.4179},
			Playlists: playlists,
			Meta:      toJSON(map[string]interface{}{"brightness": 80, "resolution": "1920x1080"}),
		},
	}
	if err := db.Create(&screens).Error; err != nil {
		return fmt.Errorf("failed to seed screens: %w", err)
	}

	return nil
}

// Helper function to convert a map to datatypes.JSON.
func toJSON(data map[string]interface{}) datatypes.JSON {
	jsonData, _ := json.Marshal(data)
	return jsonData
}
