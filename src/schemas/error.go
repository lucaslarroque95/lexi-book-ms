package schemas

// ErrorResponse is the shape of every error response returned by this API.
type ErrorResponse struct {
	Error string `json:"error"`
}
