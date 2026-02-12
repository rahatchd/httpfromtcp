package response

import (
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
)

const httpVersion = "HTTP/1.1"
const crlf = "\r\n"

type writerState int

const (
	writerStateStatusLine = iota
	writerStateHeaders
	writerStateBody
	writerStateTrailers
	writerStateDone
)

func (ws writerState) toString() string {
	switch ws {
	case writerStateStatusLine:
		return "status line"
	case writerStateHeaders:
		return "headers"
	case writerStateBody:
		return "body"
	case writerStateTrailers:
		return "trailers"
	case writerStateDone:
		return "done"
	default:
		return "unknown"
	}
}

type Writer struct {
	writer io.Writer
	state  writerState
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		writer: w,
		state:  writerStateStatusLine,
	}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.state != writerStateStatusLine {
		return fmt.Errorf("cannot write status line in state %s", w.state.toString())
	}
	_, err := fmt.Fprintf(w.writer, "%s %d %s%s", httpVersion, statusCode, statusCode.ToString(), crlf)
	if err != nil {
		return err
	}
	w.state = writerStateHeaders
	return nil
}

func writeHeaders(w io.Writer, headers headers.Headers) error {
	for k, v := range headers {
		_, err := fmt.Fprintf(w, "%s: %s%s", k, v, crlf)
		if err != nil {
			return err
		}
	}
	_, err := w.Write([]byte(crlf))
	if err != nil {
		return err
	}
	return nil
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.state != writerStateHeaders {
		return fmt.Errorf("cannot write headers in state %s", w.state.toString())
	}
	if err := writeHeaders(w.writer, headers); err != nil {
		return err
	}
	w.state = writerStateBody
	return nil
}

func (w *Writer) WriteBody(buf []byte) error {
	if w.state != writerStateBody {
		return fmt.Errorf("cannot write body in state %s", w.state.toString())
	}
	if _, err := w.writer.Write(buf); err != nil {
		return err
	}
	_, err := w.writer.Write([]byte(crlf))
	if err != nil {
		return err
	}
	w.state = writerStateTrailers
	return err
}

func (w *Writer) WriteChunkedBody(buf []byte) (int, error) {
	if w.state != writerStateBody {
		return 0, fmt.Errorf("cannot write body chunk in state %s", w.state.toString())
	}
	numBytes := len(buf)
	hex := fmt.Sprintf("%X", numBytes)
	if _, err := fmt.Fprintf(w.writer, "%s%s", hex, crlf); err != nil {
		return 0, err
	}
	if _, err := fmt.Fprintf(w.writer, "%s%s", buf, crlf); err != nil {
		return len(hex) + len(crlf), err
	}
	return len(hex) + numBytes + 2*len(crlf), nil
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	if w.state != writerStateBody {
		return 0, fmt.Errorf("cannot write body chunk done in state %s", w.state.toString())
	}
	if _, err := fmt.Fprintf(w.writer, "0%s", crlf); err != nil {
		return 0, err
	}
	w.state = writerStateTrailers
	return 1 + len(crlf), nil
}

func (w *Writer) WriteTrailers(h headers.Headers) error {
	if w.state != writerStateTrailers {
		return fmt.Errorf("cannot write trailers in state %s", w.state.toString())
	}
	if err := writeHeaders(w.writer, h); err != nil {
		return err
	}
	w.state = writerStateDone
	return nil
}
