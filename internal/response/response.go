package response

import (
	"httpfromtcp/internal/headers"
	"strconv"
)

type StatusCode int

const (
	StatusCodeOK                  = 200
	StatusCodeBadRequest          = 400
	StatusCodeInternalServerError = 500
)

func (s StatusCode) ToString() string {
	switch s {
	case StatusCodeOK:
		return "OK"
	case StatusCodeBadRequest:
		return "Bad Request"
	case StatusCodeInternalServerError:
		return "Internal Server Error"
	default:
		return ""
	}
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.NewHeaders()
	h.Set("Content-Length", strconv.Itoa(contentLen))
	h.Set("Connection", "close")
	h.Set("Content-Type", "text/plain")
	return h
}
