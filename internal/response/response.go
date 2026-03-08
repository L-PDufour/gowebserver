package response

import (
	"fmt"
	"io"
	"strconv"

	"gowebserver/internal/headers"
)

type StatusCode int

const (
	StatusOk         StatusCode = 200
	StatusBadRequest StatusCode = 400
	StatusError      StatusCode = 500
)

type HandlerError struct {
	StatusCode StatusCode
	Msg        string
}
type writerState int

const (
	writerStateStatusLine writerState = iota
	writerStateHeaders
	writerStateBody
	writerStateDone
)

type Writer struct {
	writer io.Writer
	state  writerState
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{writer: w, state: writerStateStatusLine}
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.NewHeaders()
	h.Set("Content-Length", strconv.Itoa(contentLen))
	h.Set("Connection", "close")
	h.Set("Content-Type", "text/plain")
	return h
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.state != writerStateStatusLine {
		return fmt.Errorf("cannot write status line in current state")
	}
	w.state = writerStateHeaders
	statuses := map[StatusCode]string{
		StatusOk:         "HTTP/1.1 200 OK\r\n",
		StatusBadRequest: "HTTP/1.1 400 Bad Request\r\n",
		StatusError:      "HTTP/1.1 500 Internal Server Error\r\n",
	}
	line, ok := statuses[statusCode]
	if !ok {
		return fmt.Errorf("unknown status code: %d", statusCode)
	}
	_, err := w.writer.Write([]byte(line))
	return err
}

func (w *Writer) WriteHeaders(h headers.Headers) error {
	if w.state != writerStateHeaders {
		return fmt.Errorf("cannot write headers in current state")
	}
	w.state = writerStateBody
	for key, value := range h {
		_, err := w.writer.Write([]byte(fmt.Sprintf("%s: %s\r\n", key, value)))
		if err != nil {
			return err
		}
	}
	_, err := w.writer.Write([]byte("\r\n"))
	return err
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.state != writerStateBody {
		return 0, fmt.Errorf("cannot write body in current state")
	}
	w.state = writerStateDone
	return w.writer.Write(p)
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	_, err := w.writer.Write([]byte(fmt.Sprintf("%x\r\n", len(p))))
	if err != nil {
		return 0, err
	}
	n, err := w.writer.Write([]byte(fmt.Sprintf("%s\r\n", p)))
	if err != nil {
		return n, err
	}
	return n, nil
}

func (w *Writer) WriteChunkedBodyDone() error {
	_, err := w.writer.Write([]byte("0\r\n\r\n"))
	return err
}
