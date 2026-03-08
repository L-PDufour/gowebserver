package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"gowebserver/internal/request"
	"gowebserver/internal/response"
	"gowebserver/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func handler(w *response.Writer, req *request.Request) {
	h := response.GetDefaultHeaders(0)
	h.Set("Content-Type", "text/html")

	path := strings.Split(req.RequestLine.RequestTarget, "/")

	switch path[1] {
	case "httpbin":
		remaining := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin")
		resp, err := http.Get(fmt.Sprintf("https://httpbin.org%s", remaining))
		if err != nil {
			return
		}
		defer resp.Body.Close()

		if path[2] == "stream" {
			h.Delete("Content-Length")
			h.Set("Transfer-Encoding", "chunked")
			w.WriteStatusLine(response.StatusCode(resp.StatusCode))
			w.WriteHeaders(h)

			buf := make([]byte, 1024)
			for {
				n, err := resp.Body.Read(buf)
				if n > 0 {
					w.WriteChunkedBody(buf[:n])
				}
				if err == io.EOF {
					break
				}
				if err != nil {
					return
				}
			}
			w.WriteChunkedBodyDone()
		} else {
			// regular — read full body and respond
		}
	case "yourproblem":
		body := []byte(`<html>
  <head><title>400 Bad Request</title></head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`)
		h.Set("Content-Length", strconv.Itoa(len(body)))
		w.WriteStatusLine(response.StatusBadRequest)
		w.WriteHeaders(h)
		w.WriteBody(body)
	case "myproblem":
		body := []byte(`<html>
  <head><title>500 Internal Server Error</title></head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`)
		h.Set("Content-Length", strconv.Itoa(len(body)))
		w.WriteStatusLine(response.StatusError)
		w.WriteHeaders(h)
		w.WriteBody(body)
	default:
		body := []byte(`<html>
  <head><title>200 OK</title></head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`)
		h.Set("Content-Length", strconv.Itoa(len(body)))
		w.WriteStatusLine(response.StatusOk)
		w.WriteHeaders(h)
		w.WriteBody(body)
	}
}
