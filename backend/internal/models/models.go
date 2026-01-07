package models

import (
	"database/sql"
	"encoding/json"
	"time"
)

// Driver represents registered drivers
type Driver struct {
	ID           uint           `gorm:"primaryKey;column:id" json:"id"`
	Username     string         `gorm:"column:username;size:100;uniqueIndex;not null" json:"username"`
	PasswordHash string         `gorm:"column:password_hash;size:255;not null" json:"-"`
	PhoneNumber  string         `gorm:"column:phone_number;size:50;not null" json:"phone_number"`
	FirstName    string         `gorm:"column:first_name;size:100" json:"first_name"`
	LastName     string         `gorm:"column:last_name;size:100" json:"last_name"`
	Email        sql.NullString `gorm:"column:email;size:255" json:"-"`
	IsActive     bool           `gorm:"column:is_active;default:true" json:"is_active"`
	CreatedAt    time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"column:updated_at" json:"updated_at"`
}

func (Driver) TableName() string {
	return "drivers"
}

// MarshalJSON customizes JSON serialization for Driver to handle sql.NullString properly
func (d Driver) MarshalJSON() ([]byte, error) {
	type Alias Driver
	var email *string
	if d.Email.Valid {
		email = &d.Email.String
	}
	return json.Marshal(&struct {
		Email *string `json:"email,omitempty"`
		*Alias
	}{
		Email: email,
		Alias: (*Alias)(&d),
	})
}

// Load represents loads to be assigned to drivers
type Load struct {
	ID              uint           `gorm:"primaryKey;column:id" json:"id"`
	LoadNumber      string         `gorm:"column:load_number;size:100;uniqueIndex;not null" json:"load_number"`
	DriverID        sql.NullInt64  `gorm:"column:driver_id;index" json:"driver_id"`
	Driver          *Driver        `gorm:"foreignKey:DriverID" json:"driver,omitempty"`
	Status          string         `gorm:"column:status;size:50;default:Unassigned;index" json:"status"` // Unassigned, Assigned, Completed
	Description     sql.NullString `gorm:"column:description;type:text" json:"description"`
	PickupAddress   sql.NullString `gorm:"column:pickup_address;type:text" json:"pickup_address"`
	DeliveryAddress sql.NullString `gorm:"column:delivery_address;type:text" json:"delivery_address"`
	ScheduledDate   sql.NullTime   `gorm:"column:scheduled_date" json:"scheduled_date"`
	CompletedAt     sql.NullTime   `gorm:"column:completed_at" json:"completed_at"`
	MeetingStarted  bool           `gorm:"column:meeting_started;default:false" json:"meeting_started"`
	CreatedByID     int            `gorm:"column:created_by_id;not null" json:"created_by_id"`
	CreatedAt       time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"column:updated_at" json:"updated_at"`
}

func (Load) TableName() string {
	return "loads"
}

// MarshalJSON customizes JSON serialization for Load to handle sql.NullString, sql.NullInt64, and sql.NullTime properly
func (l Load) MarshalJSON() ([]byte, error) {
	var description *string
	if l.Description.Valid {
		description = &l.Description.String
	}

	var pickupAddress *string
	if l.PickupAddress.Valid {
		pickupAddress = &l.PickupAddress.String
	}

	var deliveryAddress *string
	if l.DeliveryAddress.Valid {
		deliveryAddress = &l.DeliveryAddress.String
	}

	var driverID *uint
	if l.DriverID.Valid {
		id := uint(l.DriverID.Int64)
		driverID = &id
	}

	var scheduledDate *time.Time
	if l.ScheduledDate.Valid {
		scheduledDate = &l.ScheduledDate.Time
	}

	var completedAt *time.Time
	if l.CompletedAt.Valid {
		completedAt = &l.CompletedAt.Time
	}

	// Create a struct that excludes the sql.Null* fields and replaces them with proper types
	return json.Marshal(&struct {
		ID              uint       `json:"id"`
		LoadNumber      string     `json:"load_number"`
		DriverID        *uint      `json:"driver_id,omitempty"`
		Driver          *Driver    `json:"driver,omitempty"`
		Status          string     `json:"status"`
		Description     *string    `json:"description,omitempty"`
		PickupAddress   *string    `json:"pickup_address,omitempty"`
		DeliveryAddress *string    `json:"delivery_address,omitempty"`
		ScheduledDate   *time.Time `json:"scheduled_date,omitempty"`
		CompletedAt     *time.Time `json:"completed_at,omitempty"`
		MeetingStarted  bool       `json:"meeting_started"`
		CreatedByID     int        `json:"created_by_id"`
		CreatedAt       time.Time  `json:"created_at"`
		UpdatedAt       time.Time  `json:"updated_at"`
	}{
		ID:              l.ID,
		LoadNumber:      l.LoadNumber,
		DriverID:        driverID,
		Driver:          l.Driver,
		Status:          l.Status,
		Description:     description,
		PickupAddress:   pickupAddress,
		DeliveryAddress: deliveryAddress,
		ScheduledDate:   scheduledDate,
		CompletedAt:     completedAt,
		MeetingStarted:  l.MeetingStarted,
		CreatedByID:     l.CreatedByID,
		CreatedAt:       l.CreatedAt,
		UpdatedAt:       l.UpdatedAt,
	})
}

// Session represents user sessions (stored in database for stateless backend)
type Session struct {
	ID        uint      `gorm:"primaryKey;column:id" json:"id"`
	SessionID string    `gorm:"column:session_id;size:255;uniqueIndex;not null" json:"session_id"`
	UserID    int       `gorm:"column:user_id;not null;index" json:"user_id"`
	UserType  string    `gorm:"column:user_type;size:50;not null" json:"user_type"` // 'admin', 'driver'
	Data      string    `gorm:"column:data;type:text" json:"data"`
	ExpiresAt time.Time `gorm:"column:expires_at;not null;index" json:"expires_at"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (Session) TableName() string {
	return "sessions"
}

// MeetingRoom represents meeting rooms for video calls
type MeetingRoom struct {
	ID           uint           `gorm:"primaryKey;column:id" json:"id"`
	LoadID       uint           `gorm:"column:load_id;not null;index" json:"load_id"`
	RoomID       string         `gorm:"column:roomId;size:255;not null;uniqueIndex" json:"roomId"`
	ChannelName  string         `gorm:"column:channelName;size:255;not null;index" json:"channelName"`
	MeetingLink  string         `gorm:"column:meetingLink;size:500;not null" json:"meetingLink"`
	LoadNumber   sql.NullString `gorm:"column:load_number;size:100;index" json:"load_number"`
	SaveType     sql.NullString `gorm:"column:save_type;size:50" json:"save_type"`
	Status       string         `gorm:"column:status;type:enum('active','ended');default:active;index" json:"status"`
	CreatedAt    time.Time      `gorm:"column:created_at;index" json:"created_at"`
	LastJoinedAt sql.NullTime   `gorm:"column:lastJoinedAt" json:"lastJoinedAt"`
}

func (MeetingRoom) TableName() string {
	return "meeting_rooms"
}

// Gallery represents the gallery table for storing screenshots and recordings
type Gallery struct {
	ID                uint      `gorm:"primaryKey;column:id" json:"id"`
	LoadID            uint      `gorm:"column:load_id;index;not null" json:"load_id"`
	Load              *Load     `gorm:"foreignKey:LoadID" json:"load,omitempty"`
	FileName          string    `gorm:"column:file_name;size:500;not null" json:"file_name"`
	S3Key             string    `gorm:"column:s3_key;size:500;not null" json:"s3_key"`
	VideoRecordingKey string    `gorm:"column:video_recording_key;size:500" json:"video_recording_key"`
	CreatedAt         time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt         time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (Gallery) TableName() string {
	return "gallery"
}
