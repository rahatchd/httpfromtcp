package main

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"httpfromtcp/internal/headers"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"httpfromtcp/internal/server"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

const port = 42069

func main() {
	server, err := server.Serve(port, handlerFn)
	if err != nil {
		log.Fatalf("error starting server: %v", err)
	}
	defer server.Close()
	log.Println("server started on port", port)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("server gracefully stopped")
}

const httpbinPrefix = "/httpbin/"
const httpbinURL = "https://httpbin.org/"

func handlerFn(w *response.Writer, r *request.Request) {
	path := r.RequestLine.RequestTarget
	if after, ok := strings.CutPrefix(path, httpbinPrefix); ok {
		handleChunked(after, w)
	} else {
		handleSync(path, w)
	}
}

func handleChunked(path string, w *response.Writer) {
	url := httpbinURL + path
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("error fetching proxy: %v", err)
		handleSync("/myproblem", w)
		return
	}
	defer resp.Body.Close()

	if err := w.WriteStatusLine(response.StatusCodeOK); err != nil {
		log.Printf("error writing status line: %v", err)
		return
	}
	hdrs := response.GetDefaultHeaders(0)
	hdrs.Unset("Content-Length")
	hdrs.Set("Transfer-Encoding", "chunked")
	hdrs.Set("Trailer", "X-Content-SHA256")
	hdrs.Set("Trailer", "X-Content-Length")
	if err := w.WriteHeaders(hdrs); err != nil {
		log.Printf("error writing headers: %v", err)
		return
	}

	bufSize := 1024
	buf := make([]byte, bufSize)
	body := make([]byte, bufSize)
	numBytesRead := 0
	for {
		n, err := resp.Body.Read(buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				log.Printf("done reading proxy")
				break
			}
			log.Printf("error reading proxy: %v", err)
			break
		}
		if numBytesRead + n >= len(body) {
			newBody := make([]byte, 2 * len(body))
			copy(newBody[:numBytesRead], body[:numBytesRead])
			body = newBody
		}
		copy(body[numBytesRead:], buf[:n])
		numBytesRead += n
		log.Printf("read %d bytes from proxy (total %d)", n, numBytesRead)
		if _, err := w.WriteChunkedBody(buf[:n]); err != nil {
			log.Printf("error writing body chunk: %v", err)
			break
		}
	}

	if _, err = w.WriteChunkedBodyDone(); err != nil {
		log.Printf("error writing body chunk done: %v", err)
		return
	}
	trailers := headers.NewHeaders()
	checksum := sha256.Sum256(body[:numBytesRead])
	trailers.Set("X-Content-SHA256", fmt.Sprintf("%x", checksum))
	trailers.Set("X-Content-Length", strconv.Itoa(numBytesRead))
	if err = w.WriteTrailers(trailers); err != nil {
		log.Printf("error writing trailers: %v", err)
	}
}

func handleSync(path string, w *response.Writer) {
	res := routeSync(path)
	if err := w.WriteStatusLine(res.StatusCode); err != nil {
		log.Printf("error writing status line: %v", err)
		return
	}
	if err := w.WriteHeaders(res.headers); err != nil {
		log.Printf("error writing headers: %v", err)
		return
	}
	if err := w.WriteBody(res.body); err != nil {
		log.Printf("error writing body: %v", err)
		return
	}
	if err := w.WriteTrailers(nil); err != nil {
		log.Printf("error writing trailers: %v", err)
		return
	}
}

type responseData struct {
	StatusCode response.StatusCode
	headers headers.Headers
	body       []byte
}

const videoPath = "assets/vim.mp4"

func routeSync(path string) responseData {
	switch path {
	case "/yourproblem":
		var sc response.StatusCode = response.StatusCodeBadRequest
		body := []byte(templateSubstitution(sc, sc.ToString(), "Your request honestly kinda sucked."))
		hdrs := response.GetDefaultHeaders(len(body))
		hdrs.Override("Content-Type", "text/html")
		return responseData{
			StatusCode: sc,
			headers:    hdrs,
			body:       body,
		}
	case "/myproblem":
		var sc response.StatusCode = response.StatusCodeInternalServerError
		body := []byte(templateSubstitution(sc, sc.ToString(), "Okay, you know what? This one is on me."))
		hdrs := response.GetDefaultHeaders(len(body))
		hdrs.Override("Content-Type", "text/html")
		return responseData{
			StatusCode: sc,
			headers:    hdrs,
			body:       body,
		}
	case "/video":
		var sc response.StatusCode = response.StatusCodeOK
		body, err := os.ReadFile(videoPath)
		if err != nil {
			log.Println("error reading video (make sure you download it to the correct location): %w", err)
			return routeSync("/myproblem")
		}
		hdrs := response.GetDefaultHeaders(len(body))
		hdrs.Override("Content-Type", "video/mp4")
		return responseData{
			StatusCode: sc,
			headers:    hdrs,
			body:       body,
		}
	default:
		var sc response.StatusCode = response.StatusCodeOK
		body := []byte(templateSubstitution(sc, "Success!", "Your request was an absolute banger."))
		hdrs := response.GetDefaultHeaders(len(body))
		hdrs.Override("Content-Type", "text/html")
		return responseData{
			StatusCode: sc,
			headers:    hdrs,
			body:       body,
		}
	}
}

const htmlTemplate = `<html>
  <head>
    <title>%d %s</title>
  </head>
  <body>
    <h1>%s</h1>
	<p>%s</p>
  </body>
</html>
`

func templateSubstitution(statusCode response.StatusCode, hdr, msg string) string {
	return fmt.Sprintf(htmlTemplate,
		statusCode, statusCode.ToString(),
		hdr,
		msg,
	)
}
