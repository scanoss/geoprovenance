package models

type QuerySummary struct {
	PurlsFailedToParse []string
	PurlsWOInfo        []string
	PurlsNotFound      []string
	PurlsTooMuchData   []string
	TotalPurls         int
}
