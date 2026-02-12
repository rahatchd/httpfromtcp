package request

import (
	"bytes"
	"errors"
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
	"strconv"
	"strings"
	"unicode"
)

const (
	bufferSize = 8
	crlf       = "\r\n"
)

type requestState int

const (
	requestStateInitialized requestState = iota
	requestStateParsingHeaders
	requestStateParsingBody
	requestStateDone
)

type Request struct {
	RequestLine RequestLine
	headers.Headers
	Body  []byte
	state requestState
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(r io.Reader) (*Request, error) {
	request := &Request{state: requestStateInitialized}
	buf := make([]byte, bufferSize)
	readToIndex := 0
	for request.state != requestStateDone {
		if readToIndex >= len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}

		readBytes, err := r.Read(buf[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				if request.state != requestStateDone {
					return nil, fmt.Errorf("reader exhausted but parsing is incomplete. partial request: %v", request)
				}
				break
			}
			return nil, err
		}
		readToIndex += readBytes

		numBytesParsed, err := request.parse(buf[:readToIndex])
		if err != nil {
			return nil, err
		}

		copy(buf, buf[numBytesParsed:])
		readToIndex -= numBytesParsed
	}

	return request, nil
}

func (r *Request) parse(buf []byte) (int, error) {
	totalBytesParsed := 0
	for r.state != requestStateDone {
		n, err := r.parseSingle(buf[totalBytesParsed:])
		if err != nil {
			return 0, err
		}
		totalBytesParsed += n
		if n == 0 {
			break
		}
	}
	return totalBytesParsed, nil
}

func (r *Request) parseSingle(buf []byte) (int, error) {
	switch r.state {
	case requestStateDone:
		return 0, errors.New("error: trying to read data in a done state")
	case requestStateInitialized:
		requestLine, nb, err := parseRequestLine(buf)
		if nb > 0 && err == nil {
			r.RequestLine = *requestLine
			r.Headers = headers.NewHeaders()
			r.state = requestStateParsingHeaders
			return nb, nil
		}
		return nb, err
	case requestStateParsingHeaders:
		n, done, err := r.Headers.Parse(buf)
		if err != nil {
			return 0, fmt.Errorf("error parsing headers: %s", err)
		}
		if done {
			r.state = requestStateParsingBody
		}
		return n, nil
	case requestStateParsingBody:
		contentLength, ok := r.Headers.Get("Content-Length")
		if !ok {
			r.state = requestStateDone
			return 0, nil
		}
		r.Body = append(r.Body, buf...)
		bodyLimit, err := strconv.Atoi(contentLength)
		if err != nil {
			return 0, fmt.Errorf("invalid content length '%s': %s", contentLength, err)
		}
		if len(r.Body) < bodyLimit {
			return len(buf), nil
		}
		if len(r.Body) > bodyLimit {
			return 0, fmt.Errorf("error: body (%d) larger than content-length (%d)", len(r.Body), bodyLimit)
		}
		r.state = requestStateDone
		return len(buf), nil
	default:
		return 0, fmt.Errorf("error: unknown parser state %d", r.state)
	}
}

func parseRequestLine(b []byte) (*RequestLine, int, error) {
	before, _, ok := bytes.Cut(b, []byte(crlf))
	if !ok {
		return nil, 0, nil
	}
	numBytesParsed := len(before)
	reqLine, err := requestLineFromString(string(before))
	return reqLine, numBytesParsed + 2, err
}

func requestLineFromString(s string) (*RequestLine, error) {
	parts := strings.Split(s, " ")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid number of parts in request-line: %s", s)
	}

	method := parts[0]
	for _, c := range method {
		if !unicode.IsLetter(c) || !unicode.IsUpper(c) {
			return nil, fmt.Errorf("invalid method: %s", method)
		}
	}

	requestTarget := parts[1]

	httpName := strings.Split(parts[2], "/")
	if len(httpName) != 2 || httpName[0] != "HTTP" {
		return nil, fmt.Errorf("invalid version: %s", httpName[0])
	}
	version := strings.Split(httpName[1], ".")
	if len(version) != 2 {
		return nil, fmt.Errorf("invalid version: %s", httpName[1])
	}
	for _, v := range version {
		for _, c := range v {
			if !unicode.IsDigit(c) {
				return nil, fmt.Errorf("invalid version: %s", httpName[1])
			}
		}
		if v != "1" {
			return nil, fmt.Errorf("unsupported version: %s", httpName[1])
		}
	}

	return &RequestLine{
		HttpVersion:   "1.1",
		RequestTarget: requestTarget,
		Method:        method,
	}, nil
}
