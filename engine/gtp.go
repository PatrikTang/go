package engine

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"sync"
)

// GTPClient is a synchronous GTP client over an arbitrary reader/writer pair.
type GTPClient struct {
	r      *bufio.Reader
	w      io.Writer
	closer io.Closer
	mu     sync.Mutex
}

func NewGTPClient(r io.Reader, w io.Writer, closer io.Closer) *GTPClient {
	return &GTPClient{r: bufio.NewReader(r), w: w, closer: closer}
}

func (c *GTPClient) Close() error {
	if c.closer != nil {
		return c.closer.Close()
	}
	return nil
}

func (c *GTPClient) Command(cmd string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, err := io.WriteString(c.w, cmd+"\n"); err != nil {
		return "", err
	}
	return c.readResponse()
}

func (c *GTPClient) readResponse() (string, error) {
	var body []string
	success := true
	first := true
	for {
		line, err := c.r.ReadString('\n')
		if err != nil {
			if err == io.EOF && len(body) > 0 {
				break
			}
			return "", err
		}
		line = strings.TrimRight(line, "\r\n")
		if first {
			first = false
			switch {
			case strings.HasPrefix(line, "="):
				line = strings.TrimSpace(strings.TrimPrefix(line, "="))
			case strings.HasPrefix(line, "?"):
				success = false
				line = strings.TrimSpace(strings.TrimPrefix(line, "?"))
			default:
				return "", fmt.Errorf("malformed GTP response: %q", line)
			}
			body = append(body, line)
			continue
		}
		if line == "" {
			break
		}
		body = append(body, line)
	}
	out := strings.Join(body, "\n")
	if !success {
		return out, fmt.Errorf("gtp error: %s", out)
	}
	return out, nil
}

// Stream sends a command and invokes handler for each non-empty output line
// until handler returns false or the engine emits a terminating response.
// When handler returns false, Stream sends a harmless command to interrupt
// the engine's output and drains the response.
func (c *GTPClient) Stream(cmd string, handler func(line string) bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, err := io.WriteString(c.w, cmd+"\n"); err != nil {
		return err
	}
	for {
		line, err := c.r.ReadString('\n')
		if err != nil {
			return err
		}
		line = strings.TrimRight(line, "\r\n")
		if strings.HasPrefix(line, "=") || strings.HasPrefix(line, "?") {
			for {
				l, err := c.r.ReadString('\n')
				if err != nil {
					return err
				}
				if strings.TrimRight(l, "\r\n") == "" {
					return nil
				}
			}
		}
		if line == "" {
			continue
		}
		if !handler(line) {
			if _, err := io.WriteString(c.w, "name\n"); err != nil {
				return err
			}
			for {
				l, err := c.r.ReadString('\n')
				if err != nil {
					return err
				}
				l = strings.TrimRight(l, "\r\n")
				if strings.HasPrefix(l, "=") || strings.HasPrefix(l, "?") {
					for {
						l2, err := c.r.ReadString('\n')
						if err != nil {
							return err
						}
						if strings.TrimRight(l2, "\r\n") == "" {
							return nil
						}
					}
				}
			}
		}
	}
}
