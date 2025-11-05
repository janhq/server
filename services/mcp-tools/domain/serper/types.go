package serper

// TBSTimeRange defines time-based search filters for Serper API
type TBSTimeRange string

const (
	TBSAny       TBSTimeRange = ""
	TBSPastHour  TBSTimeRange = "qdr:h"
	TBSPastDay   TBSTimeRange = "qdr:d"
	TBSPastWeek  TBSTimeRange = "qdr:w"
	TBSPastMonth TBSTimeRange = "qdr:m"
	TBSPastYear  TBSTimeRange = "qdr:y"
)

// SearchRequest represents a search query to Serper API
type SearchRequest struct {
	Q           string        `json:"q"`
	GL          *string       `json:"gl,omitempty"`          // Region code (ISO 3166-1 alpha-2)
	HL          *string       `json:"hl,omitempty"`          // Language code (ISO 639-1)
	Location    *string       `json:"location,omitempty"`    // Location for search results
	Num         *int          `json:"num,omitempty"`         // Number of results (default: 10)
	Page        *int          `json:"page,omitempty"`        // Page number (default: 1)
	Autocorrect *bool         `json:"autocorrect,omitempty"` // Enable autocorrect
	TBS         *TBSTimeRange `json:"tbs,omitempty"`         // Time-based search filter
}

// SearchResponse contains search results from Serper API
type SearchResponse struct {
	SearchParameters map[string]any   `json:"searchParameters"`
	Organic          []map[string]any `json:"organic"`
	KnowledgeGraph   map[string]any   `json:"knowledgeGraph,omitempty"`
	Images           []map[string]any `json:"images,omitempty"`
	News             []map[string]any `json:"news,omitempty"`
	AnswerBox        map[string]any   `json:"answerBox,omitempty"`
}

// FetchWebpageRequest represents a webpage scraping request
type FetchWebpageRequest struct {
	Url             string `json:"url"`
	IncludeMarkdown *bool  `json:"includeMarkdown,omitempty"`
}

// FetchWebpageResponse contains scraped webpage content
type FetchWebpageResponse struct {
	Text     string         `json:"text"`
	Metadata map[string]any `json:"metadata"`
}
