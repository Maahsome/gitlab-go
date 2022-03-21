package gitlab

import "flag"

type PaginationOptions struct {
	Page    int `url:"page,omitempty"`
	PerPage int `url:"per_page,omitempty"`
}

type SortDirection string

const (
	SortDirectionAsc  SortDirection = "asc"
	SortDirectionDesc SortDirection = "desc"
)

type SortOptions struct {
	OrderBy string        `url:"order_by,omitempty"`
	Sort    SortDirection `url:"sort,omitempty"`
}

type ResponseWithMessage struct {
	Message string `json:"message"`
}

type ResponseMeta struct {
	Method     string
	Url        string
	StatusCode int
	RequestId  string
	Page       int
	PerPage    int
	PrevPage   int
	NextPage   int
	TotalPages int
	Total      int
	Runtime    float64
}

type Message struct {
	Message string `json:"message"`
}

const (
	dateLayout = "2006-01-02T15:04:05-07:00"
)

var (
	skipCertVerify = flag.Bool("gitlab.skip-cert-check", false,
		`If set to true, gitlab client will skip certificate checking for https, possibly exposing your system to MITM attack.`)
)
