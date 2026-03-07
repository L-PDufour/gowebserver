package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"

	"gowebserver/internal/headers"
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	state       int
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

const (
	requestStateInitialized = iota
	requestStateParsingHeaders
	requestStateDone
)

var (
	bufferSize            = 8
	ErrInvalidRequestLine = errors.New("invalid request line")
)

func NewRequest() *Request {
	return &Request{
		RequestLine: RequestLine{},
		Headers:     headers.NewHeaders(),
		state:       requestStateInitialized,
	}
}

func (r RequestLine) String() string {
	return fmt.Sprintf("Request line:\n- Method: %s\n- Target: %s\n- Version: %s", r.Method, r.RequestTarget, r.HttpVersion)
}

func IsLetter(s string) bool {
	return !strings.ContainsFunc(s, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsUpper(r)
	})
}

func (r *Request) parse(data []byte) (int, error) {
	totalBytesParsed := 0
	for r.state != requestStateDone {
		n, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return 0, err
		}
		if n == 0 {
			break
		}
		totalBytesParsed += n
	}
	return totalBytesParsed, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.state {
	case requestStateInitialized: // ← parse when initialized
		line, n, err := parseRequestLine(data)
		if err != nil {
			return n, err
		}
		if n == 0 {
			return 0, nil
		}
		r.state = requestStateParsingHeaders // ← transition to done
		r.RequestLine = *line
		return n, nil
	case requestStateDone: // ← error when already done
		return 0, fmt.Errorf("error: trying to read data in a done state")
	case requestStateParsingHeaders:
		n, done, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}
		if done {
			r.state = requestStateDone
		}
		return n, nil
	default:
		return 0, fmt.Errorf("error: unknown state")
	}
}

func requestLineFromString(str string) (*RequestLine, error) {
	parts := strings.Split(str, " ")
	if len(parts) != 3 {
		return nil, fmt.Errorf("poorly formatted request-line: %s", str)
	}
	method := parts[0]
	if !IsLetter(method) {
		return nil, fmt.Errorf("invalid method: %s", method)
	}

	requestTarget := parts[1]

	versionParts := strings.Split(parts[2], "/")
	if len(versionParts) != 2 {
		return nil, fmt.Errorf("malformed start-line: %s", str)
	}

	httpPart := versionParts[0]
	if httpPart != "HTTP" {
		return nil, fmt.Errorf("unrecognized HTTP-version: %s", httpPart)
	}
	version := versionParts[1]
	if version != "1.1" {
		return nil, fmt.Errorf("unrecognized HTTP-version: %s", version)
	}

	return &RequestLine{
		Method:        method,
		RequestTarget: requestTarget,
		HttpVersion:   versionParts[1],
	}, nil
}

func parseRequestLine(data []byte) (*RequestLine, int, error) {
	idx := bytes.Index(data, []byte("\r\n"))
	if idx == -1 {
		return nil, 0, nil
	}
	requestLineText := string(data[:idx])
	requestLine, err := requestLineFromString(requestLineText)
	if err != nil {
		return nil, 0, err
	}

	return requestLine, idx + 2, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	readToIndex := 0
	request := NewRequest()

	buf := make([]byte, bufferSize)
	for request.state != requestStateDone {
		if readToIndex >= len(buf) {
			bufCopy := make([]byte, len(buf)*2)
			copy(bufCopy, buf)
			buf = bufCopy

		}
		n, err := reader.Read(buf[readToIndex:])
		if err != nil {
			if err == io.EOF {
				if request.state != requestStateDone {
					return nil, fmt.Errorf("incomplete request: unexpected EOF")
				}
				break
			}
			return nil, err
		}
		readToIndex += n
		nParsed, err := request.parse(buf[:readToIndex])
		if err != nil {
			return nil, err
		}
		copy(buf, buf[nParsed:])
		readToIndex -= nParsed
	}
	return request, nil
}
