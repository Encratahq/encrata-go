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
	ID          string `json:"id"`
	Name        string `json:"name"`
	ListType    string `json:"type"`
	Type        string `json:"list_type"`
	EmailCount  int    `json:"email_count"`
	TargetCount int    `json:"target_count"`
	CreatedAt   string `json:"created_at"`
}

// RunTrigger is the response to triggering an immediate monitoring run.
type RunTrigger struct {
	RunID   string `json:"run_id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

// RawObject holds API response blocks whose shape can vary by source.
type RawObject map[string]any

// IPInfo is an IP intelligence result.
type IPInfo struct {
	Query    string    `json:"query"`
	IP       string    `json:"ip"`
	Location RawObject `json:"location"`
	ASN      RawObject `json:"asn"`
	Company  RawObject `json:"company"`
	Threat   RawObject `json:"threat"`
	Credits  float64   `json:"credits"`
}

// PhoneInfo is a phone intelligence result.
type PhoneInfo struct {
	Query        string    `json:"query"`
	Phone        string    `json:"phone"`
	Valid        bool      `json:"valid"`
	Location     string    `json:"location"`
	Type         string    `json:"type"`
	Format       RawObject `json:"format"`
	Country      RawObject `json:"country"`
	Carrier      RawObject `json:"carrier"`
	Messaging    RawObject `json:"messaging"`
	Validation   RawObject `json:"validation"`
	Registration RawObject `json:"registration"`
	Risk         RawObject `json:"risk"`
	Breaches     RawObject `json:"breaches"`
	Credits      float64   `json:"credits"`
}

// DomainInfo is a domain intelligence result.
type DomainInfo struct {
	Domain      string    `json:"domain"`
	Whois       RawObject `json:"whois"`
	DNS         RawObject `json:"dns"`
	SSL         RawObject `json:"ssl"`
	ThreatIntel RawObject `json:"threat_intel"`
	Intel       RawObject `json:"intel"`
	Company     RawObject `json:"company"`
	Report      RawObject `json:"report"`
	Extras      RawObject `json:"extras"`
	Credits     float64   `json:"credits"`
}

// CompanyResult is one person found by company search.
type CompanyResult struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Role     string `json:"role"`
	LinkedIn string `json:"linkedin"`
}

// CompanyInfo is a company search response.
type CompanyInfo struct {
	Company string          `json:"company"`
	Results []CompanyResult `json:"results"`
	Profile RawObject       `json:"profile"`
	Total   int             `json:"total"`
	Credits float64         `json:"credits"`
}

// GoogleResult is one Google search result.
type GoogleResult struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"snippet"`
}

// GoogleSearch is a Google dork search response.
type GoogleSearch struct {
	Query      string         `json:"query"`
	Results    []GoogleResult `json:"results"`
	Enrichment RawObject      `json:"enrichment"`
	Total      int            `json:"total"`
	Credits    float64        `json:"credits"`
}

// DarkWebResult is one dark web intelligence hit.
type DarkWebResult struct {
	Source      string    `json:"source"`
	Title       string    `json:"title"`
	Body        string    `json:"body"`
	Domain      string    `json:"domain"`
	Hackishness float64   `json:"hackishness"`
	CrawlDate   string    `json:"crawl_date"`
	Emails      []string  `json:"emails"`
	IPs         []string  `json:"ips"`
	CVEs        []string  `json:"cves"`
	Leak        RawObject `json:"leak"`
	Chat        RawObject `json:"chat"`
	IRC         RawObject `json:"irc"`
	Paste       RawObject `json:"paste"`
	Forum       RawObject `json:"forum"`
	Market      RawObject `json:"market"`
}

// DarkWebSearch is a dark web search response.
type DarkWebSearch struct {
	Query       string          `json:"query"`
	Total       int             `json:"total"`
	ResultCount int             `json:"result_count"`
	Credits     float64         `json:"credits"`
	Results     []DarkWebResult `json:"results"`
}

// ScrapeResult is scraped page content as markdown plus metadata.
type ScrapeResult struct {
	Success    bool      `json:"success"`
	URL        string    `json:"url"`
	StatusCode int       `json:"status_code"`
	Content    string    `json:"content"`
	Metadata   RawObject `json:"metadata"`
	Credits    float64   `json:"credits"`
	LatencyMS  int       `json:"latency_ms"`
}

// ExtractResult is extracted page data.
type ExtractResult struct {
	Success    bool            `json:"success"`
	URL        string          `json:"url"`
	StatusCode int             `json:"status_code"`
	Extracted  json.RawMessage `json:"extracted"`
	Metadata   RawObject       `json:"metadata"`
	ErrorCode  string          `json:"error_code"`
	Error      string          `json:"error"`
	Credits    float64         `json:"credits"`
	LatencyMS  int             `json:"latency_ms"`
}

// ScreenshotResult is a captured page screenshot as base64 image data.
type ScreenshotResult struct {
	Success    bool      `json:"success"`
	URL        string    `json:"url"`
	StatusCode int       `json:"status_code"`
	Screenshot string    `json:"screenshot"`
	Format     string    `json:"format"`
	Metadata   RawObject `json:"metadata"`
	ErrorCode  string    `json:"error_code"`
	Error      string    `json:"error"`
	Credits    float64   `json:"credits"`
	LatencyMS  int       `json:"latency_ms"`
}

// FaceMatch is a face search match.
type FaceMatch struct {
	UUID        string  `json:"uuid"`
	Name        string  `json:"name"`
	Probability float64 `json:"probability"`
	Left        int     `json:"left"`
	Top         int     `json:"top"`
	Right       int     `json:"right"`
	Bottom      int     `json:"bottom"`
}

// FaceSearch is a face search response.
type FaceSearch struct {
	ImageURL      string      `json:"image_url"`
	Matched       bool        `json:"matched"`
	Threshold     float64     `json:"threshold"`
	FacesDetected int         `json:"faces_detected"`
	Matches       []FaceMatch `json:"matches"`
	Credits       float64     `json:"credits"`
	LatencyMS     int         `json:"latency_ms"`
}

// BulkSearchResponse collects streamed bulk search results.
type BulkSearchResponse struct {
	Results     []RawObject `json:"results"`
	CreditsUsed int         `json:"credits_used"`
}

// APIKey is an account API key. Key is only populated on creation.
type APIKey struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	KeyPreview    string `json:"key_preview"`
	Key           string `json:"key"`
	TotalRequests int    `json:"total_requests"`
	CreditsUsed   int    `json:"credits_used"`
	CreatedAt     string `json:"created_at"`
	LastUsed      string `json:"last_used"`
}

// Webhook is a webhook endpoint. Secret is only populated on creation.
type Webhook struct {
	ID          string   `json:"id"`
	WorkspaceID string   `json:"workspace_id"`
	URL         string   `json:"url"`
	Events      []string `json:"events"`
	IsActive    bool     `json:"is_active"`
	Description string   `json:"description"`
	Secret      string   `json:"secret"`
	CreatedBy   string   `json:"created_by"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
}

// WebhookDelivery is one webhook delivery attempt.
type WebhookDelivery struct {
	ID             string    `json:"id"`
	WebhookID      string    `json:"webhook_id"`
	EventType      string    `json:"event_type"`
	Payload        RawObject `json:"payload"`
	Status         string    `json:"status"`
	Attempts       int       `json:"attempts"`
	ResponseStatus int       `json:"response_status"`
	ResponseBody   string    `json:"response_body"`
	LastAttemptAt  string    `json:"last_attempt_at"`
	CreatedAt      string    `json:"created_at"`
}

// Workflow is an automation workflow definition.
type Workflow struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Status      string         `json:"status"`
	Trigger     RawObject      `json:"trigger"`
	Steps       []RawObject    `json:"steps"`
	TemplateID  string         `json:"template_id"`
	Version     int            `json:"version"`
	CreatedAt   string         `json:"created_at"`
	UpdatedAt   string         `json:"updated_at"`
	Extra       map[string]any `json:"-"`
}

// WorkflowRun is one workflow execution.
type WorkflowRun struct {
	ID          string      `json:"id"`
	WorkflowID  string      `json:"workflow_id"`
	Status      string      `json:"status"`
	Steps       []RawObject `json:"steps"`
	CreditsUsed int         `json:"credits_used"`
	StartedAt   string      `json:"started_at"`
	CompletedAt string      `json:"completed_at"`
}

// WorkflowTemplate is a reusable workflow template.
type WorkflowTemplate struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Category    string      `json:"category"`
	Trigger     RawObject   `json:"trigger"`
	Steps       []RawObject `json:"steps"`
}

// WorkflowSecret is a workflow secret reference. Values are never returned.
type WorkflowSecret struct {
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
