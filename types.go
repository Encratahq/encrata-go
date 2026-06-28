package encrata

import "encoding/json"

// Socials holds a person's social media profiles. Empty fields are omitted.
type Socials struct {
	LinkedIn  string `json:"linkedin,omitempty"`
	Twitter   string `json:"twitter,omitempty"`
	Instagram string `json:"instagram,omitempty"`
	Facebook  string `json:"facebook,omitempty"`
	GitHub    string `json:"github,omitempty"`
}

// BreachInfo summarizes data-breach exposure for a person.
type BreachInfo struct {
	Count       int      `json:"count"`
	Services    []string `json:"services"`
	ExposedData []string `json:"exposed_data"`
}

// RegisteredServices lists services where an email is registered.
type RegisteredServices struct {
	Count    int      `json:"count"`
	Services []string `json:"services"`
}

// NewsArticle is a news mention of a person.
type NewsArticle struct {
	Title  string `json:"title"`
	URL    string `json:"url"`
	Date   string `json:"date"`
	Source string `json:"source"`
}

// Publication is an academic publication attributed to a person.
type Publication struct {
	Title     string `json:"title"`
	URL       string `json:"url"`
	Year      *int   `json:"year"`
	Citations int    `json:"citations"`
}

// Person is the complete intelligence result of an email lookup. Nested fields
// are nil when the API returns no data for them.
type Person struct {
	Name               string              `json:"name"`
	Email              string              `json:"email"`
	Company            string              `json:"company"`
	Role               string              `json:"role"`
	Industry           string              `json:"industry"`
	Location           string              `json:"location"`
	Birthplace         string              `json:"birthplace"`
	CurrentLocation    string              `json:"current_location"`
	Bio                string              `json:"bio"`
	Age                string              `json:"age"`
	Gender             string              `json:"gender"`
	Education          string              `json:"education"`
	Phone              string              `json:"phone"`
	PhotoURL           string              `json:"photo"`
	Validity           string              `json:"validity"`
	Socials            *Socials            `json:"socials"`
	Breaches           *BreachInfo         `json:"breaches"`
	RegisteredServices *RegisteredServices `json:"registered_services"`
	News               []NewsArticle       `json:"news"`
	Publications       []Publication       `json:"publications"`
}

// Validation is the result of an email validation check.
type Validation struct {
	Email    string `json:"email"`
	Validity string `json:"validity"`
	Message  string `json:"message"`
}

// BreachReport is the standalone result of a breach check.
type BreachReport struct {
	Email       string   `json:"email"`
	Count       int      `json:"count"`
	Services    []string `json:"services"`
	ExposedData []string `json:"exposed_data"`
	Message     string   `json:"message"`
}

// Monitor is a monitoring configuration.
type Monitor struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	Status          string   `json:"status"`
	Frequency       string   `json:"frequency"`
	ChangeDetection string   `json:"change_detection"`
	DataSourceType  string   `json:"data_source_type"`
	DataSourceRef   string   `json:"data_source_ref"`
	EmailCount      int      `json:"email_count"`
	TrackedFields   []string `json:"tracked_fields"`
	LastRunAt       string   `json:"last_run_at"`
	NextRunAt       string   `json:"next_run_at"`
	CreatedAt       string   `json:"created_at"`
}

// MonitorRun is a single execution of a monitor.
type MonitorRun struct {
	ID              string `json:"id"`
	MonitorID       string `json:"monitor_id"`
	MonitorName     string `json:"monitor_name"`
	Status          string `json:"status"`
	TotalRecords    int    `json:"total_records"`
	ChangesDetected int    `json:"changes_detected"`
	CreditsUsed     int    `json:"credits_used"`
	StartedAt       string `json:"started_at"`
	CompletedAt     string `json:"completed_at"`
}

// MonitorSnapshot is one enrichment result from a monitoring run. Changes and
// Data are raw JSON; unmarshal them into a concrete type as needed.
type MonitorSnapshot struct {
	ID         string          `json:"id"`
	Email      string          `json:"email"`
	HasChanges bool            `json:"has_changes"`
	Changes    json.RawMessage `json:"changes"`
	Data       json.RawMessage `json:"data"`
}

// ContactList is a named collection of emails.
type ContactList struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	EmailCount int    `json:"email_count"`
	CreatedAt  string `json:"created_at"`
}

// RunTrigger is the response to triggering an immediate monitoring run.
type RunTrigger struct {
	RunID   string `json:"run_id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}
