package entities

import "time"

// ActivityEventType represents the type of an activity log entry.
type ActivityEventType string

const (
	ActivityEventCheckin  ActivityEventType = "checkin"
	ActivityEventCheckout ActivityEventType = "checkout"
)

// ActivityLogJukir holds jukir information for an activity log entry.
type ActivityLogJukir struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	JukirCode string `json:"jukir_code"`
}

// ActivityLogArea holds parking area information for an activity log entry.
type ActivityLogArea struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

// ActivityLogItem represents a single activity event (checkin/checkout).
type ActivityLogItem struct {
	EventTime   time.Time         `json:"event_time"`
	EventType   ActivityEventType `json:"event_type"`
	SessionID   *uint             `json:"session_id,omitempty"`
	PlatNomor   *string           `json:"plat_nomor,omitempty"`
	VehicleType string            `json:"vehicle_type"`
	IsManual    bool              `json:"is_manual"`
	Jukir       *ActivityLogJukir `json:"jukir,omitempty"`
	Area        *ActivityLogArea  `json:"area,omitempty"`
}

// ActivityLogMeta contains pagination and filter metadata for activity logs.
type ActivityLogMeta struct {
	Total     int    `json:"total"`
	Limit     int    `json:"limit"`
	Offset    int    `json:"offset"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	JukirID   *uint  `json:"jukir_id,omitempty"`
	AreaID    *uint  `json:"area_id,omitempty"`
}

// ActivityLogResponse represents the response payload for activity log requests.
type ActivityLogResponse struct {
	Activities []ActivityLogItem `json:"activities"`
	Meta       ActivityLogMeta   `json:"meta"`
}
