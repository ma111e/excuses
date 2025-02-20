package types

type FetchQuoteRequest struct {
	Path string
}

type FetchQuoteResponse struct {
	Quote        string
	NextLink     string
	PreviousLink string
	Error        string
}

type QuoteService struct{}
