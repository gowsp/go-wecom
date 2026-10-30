package longconn

import (
	"bufio"
	"bytes"
	"errors"
	"github.com/gobwas/ws"
	"net"
	"sync"
	"time"
)

// connection carries the buffered reader left over from the handshake and
// serializes frame writes. It implements io.ReadWriter so wsutil can answer
// control frames.
type connection struct {
	net.Conn
	reader       *bufio.Reader
	writeTimeout time.Duration
	writeMu      sync.Mutex
}

func (c *connection) Read(p []byte) (int, error) {
	return c.reader.Read(p)
}

// Write is used by wsutil for protocol control frames. Business frames use
// writeText, which assembles the whole frame before taking the same lock.
func (c *connection) Write(p []byte) (int, error) {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	if err := c.Conn.SetWriteDeadline(time.Now().Add(c.writeTimeout)); err != nil {
		return 0, err
	}
	return c.Conn.Write(p)
}

func (c *connection) writeText(payload []byte) error {
	frame := ws.MaskFrameInPlace(ws.NewTextFrame(payload))
	var buffer bytes.Buffer
	if err := ws.WriteFrame(&buffer, frame); err != nil {
		return errors.New("wecom: encode frame failed")
	}
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	if err := c.Conn.SetWriteDeadline(time.Now().Add(c.writeTimeout)); err != nil {
		return err
	}
	_, err := c.Conn.Write(buffer.Bytes())
	return err
}
