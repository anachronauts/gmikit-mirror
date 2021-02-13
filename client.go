package gmikit

import (
	"bufio"
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"strconv"
	"time"
)

var (
	ErrInvalidURL      = errors.New("gemini: invalid URL")
	ErrInvalidStatus   = errors.New("gemini: invalid status")
	ErrMetaTooLong     = errors.New("gemini: meta too long")
	ErrMalformedHeader = errors.New("gemini: malformed header")
)

var crlf = []byte("\r\n")

type Status int

const (
	StatusInput                    Status = 10
	StatusSensitiveInput           Status = 11
	StatusSuccess                  Status = 20
	StatusRedirect                 Status = 30
	StatusPermanentRedirect        Status = 31
	StatusTemporaryFailure         Status = 40
	StatusServerUnavailable        Status = 41
	StatusCGIError                 Status = 42
	StatusProxyError               Status = 43
	StatusSlowDown                 Status = 44
	StatusPermanentFailure         Status = 50
	StatusNotFound                 Status = 51
	StatusGone                     Status = 52
	StatusProxyRequestRefused      Status = 53
	StatusBadRequest               Status = 59
	StatusCertificateRequired      Status = 60
	StatusCertificateNotAuthorized Status = 61
	StatusCertificateNotValid      Status = 62
)

type StatusClass int

const (
	StatusClassInput               StatusClass = 1
	StatusClassSuccess             StatusClass = 2
	StatusClassRedirect            StatusClass = 3
	StatusClassTemporaryFailure    StatusClass = 4
	StatusClassPermanentFailure    StatusClass = 5
	StatusClassCertificateRequired StatusClass = 6
)

var statusStrings = map[Status]string{
	// 1x
	StatusInput:          "INPUT",
	StatusSensitiveInput: "SENSITIVE INPUT",

	// 2x
	StatusSuccess: "SUCCESS",

	// 3x
	StatusRedirect:          "REDIRECT - TEMPORARY",
	StatusPermanentRedirect: "REDIRECT - PERMANENT",

	// 4x
	StatusTemporaryFailure:  "TEMPORARY FAILURE",
	StatusServerUnavailable: "SERVER UNAVAILABLE",
	StatusCGIError:          "CGI ERROR",
	StatusProxyError:        "PROXY ERROR",
	StatusSlowDown:          "SLOW DOWN",

	// 5x
	StatusPermanentFailure:    "PERMANENT FAILURE",
	StatusNotFound:            "NOT FOUND",
	StatusGone:                "GONE",
	StatusProxyRequestRefused: "PROXY REQUEST REFUSED",
	StatusBadRequest:          "BAD REQUEST",

	// 6x
	StatusCertificateRequired:      "CLIENT CERTIFICATE REQUIRED",
	StatusCertificateNotAuthorized: "CERTIFICATE NOT AUTHORIZED",
	StatusCertificateNotValid:      "CERTIFICATE NOT VALID",
}

func (s Status) Class() StatusClass {
	return StatusClass(s / 10)
}

func (s Status) String() string {
	str, ok := statusStrings[s]
	if !ok {
		str, ok = statusStrings[Status(s.Class()*10)]
	}
	if !ok {
		str = ""
	}
	return fmt.Sprintf("%d %s", s, str)
}

type Request struct {
	URL         *url.URL
	Certificate *tls.Certificate
	Context     context.Context
	Host        string
}

func NewRequest(url *url.URL) *Request {
	req := &Request{
		URL:  url,
		Host: url.Host,
	}
	if url.Port() == "" {
		req.Host = fmt.Sprintf("%s:1965", url.Host)
	}
	return req
}

func (r *Request) Write(w *bufio.Writer) error {
	url := r.URL.String()
	if r.URL.User != nil || len(url) > 1024 {
		return ErrInvalidURL
	}
	if _, err := w.WriteString(url); err != nil {
		return err
	}
	if _, err := w.Write(crlf); err != nil {
		return err
	}
	return nil
}

type Response struct {
	Status Status
	Meta   string
	Body   io.Reader
	TLS    tls.ConnectionState
	closer io.Closer
}

func ReadResponse(rc io.ReadCloser) (*Response, error) {
	resp := &Response{}
	br := bufio.NewReader(rc)

	statusB := make([]byte, 2)
	if _, err := br.Read(statusB); err != nil {
		return nil, err
	}
	status, err := strconv.Atoi(string(statusB))
	if err != nil {
		return nil, err
	}
	if status < 10 || status >= 70 {
		return nil, ErrInvalidStatus
	}
	resp.Status = Status(status)

	if b, err := br.ReadByte(); err != nil {
		return nil, err
	} else if b != ' ' {
		return nil, ErrMalformedHeader
	}

	meta, err := br.ReadString('\r')
	if err != nil {
		return nil, err
	}
	meta = meta[:len(meta)-1]
	if len(meta) > 1024 {
		return nil, ErrMetaTooLong
	}
	if resp.Status.Class() == StatusClassSuccess && meta == "" {
		meta = "text/gemini; charset=utf-8"
	}
	resp.Meta = meta

	if b, err := br.ReadByte(); err != nil {
		return nil, err
	} else if b != '\n' {
		return nil, ErrMalformedHeader
	}

	if resp.Status.Class() != StatusClassSuccess {
		rc.Close()
	} else {
		resp.Body = br
		resp.closer = rc
	}

	return resp, nil
}

func (r *Response) Close() error {
	if r.closer != nil {
		if err := r.closer.Close(); err != nil {
			return err
		}
		r.closer = nil
	}
	return nil
}

type Client struct {
	TrustCertificate func(hostname string, cert *x509.Certificate) error
	Timeout          time.Duration
}

func (c *Client) Do(req *Request) (*Response, error) {
	config := &tls.Config{
		InsecureSkipVerify: true,
		MinVersion:         tls.VersionTLS12,
		GetClientCertificate: func(_ *tls.CertificateRequestInfo) (*tls.Certificate, error) {
			if req.Certificate != nil {
				return req.Certificate, nil
			} else {
				return &tls.Certificate{}, nil
			}
		},
		VerifyConnection: func(cs tls.ConnectionState) error {
			if c.TrustCertificate != nil {
				cert := cs.PeerCertificates[0]
				return c.TrustCertificate(req.URL.Host, cert)
			} else {
				return nil
			}
		},
		ServerName: req.URL.Host,
	}

	ctx := req.Context
	if ctx == nil {
		ctx = context.Background()
	}

	start := time.Now()
	dialer := net.Dialer{
		Timeout: c.Timeout,
	}

	netConn, err := dialer.DialContext(ctx, "tcp", req.Host)
	if err != nil {
		return nil, err
	}

	conn := tls.Client(netConn, config)
	if c.Timeout != 0 {
		err := conn.SetDeadline(start.Add(c.Timeout))
		if err != nil {
			return nil, fmt.Errorf("failed to set connection deadline: %w", err)
		}
	}

	resp, err := c.do(conn, req)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}

	resp.TLS = conn.ConnectionState()

	return resp, nil
}

func (c *Client) do(conn *tls.Conn, req *Request) (*Response, error) {
	w := bufio.NewWriter(conn)
	err := req.Write(w)
	if err != nil {
		return nil, fmt.Errorf("failed to write request: %w", err)
	}

	if err := w.Flush(); err != nil {
		return nil, err
	}

	resp, err := ReadResponse(conn)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
