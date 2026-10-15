package main

import (
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/url"
	"strings"
	"time"

	"go.bug.st/serial"
)

const culDialTimeout = 10 * time.Second

// CulConn is the transport used to talk to a CUL (USB serial or network TCP).
type CulConn interface {
	io.Reader
	io.Writer
	io.Closer
	ResetInputBuffer() error
}

type serialCulConn struct {
	port serial.Port
}

func (c *serialCulConn) Read(p []byte) (int, error)  { return c.port.Read(p) }
func (c *serialCulConn) Write(p []byte) (int, error) { return c.port.Write(p) }
func (c *serialCulConn) Close() error                { return c.port.Close() }
func (c *serialCulConn) ResetInputBuffer() error     { return c.port.ResetInputBuffer() }

type tcpCulConn struct {
	conn net.Conn
}

func (c *tcpCulConn) Read(p []byte) (int, error)  { return c.conn.Read(p) }
func (c *tcpCulConn) Write(p []byte) (int, error) { return c.conn.Write(p) }
func (c *tcpCulConn) Close() error                { return c.conn.Close() }

// ResetInputBuffer is a no-op for TCP; there is no UART RX buffer to purge.
func (c *tcpCulConn) ResetInputBuffer() error { return nil }

// openCulConn opens the configured CUL transport.
// serial_port may be a device path (e.g. /dev/ttyUSB0) or a URL:
// tcp://host:port or telnet://host:port (alias; raw TCP, baud unused).
func openCulConn() CulConn {
	addr := strings.TrimSpace(config.SerialPort)
	lower := strings.ToLower(addr)

	if strings.HasPrefix(lower, "tcp://") || strings.HasPrefix(lower, "telnet://") {
		return openTCPCul(addr)
	}
	return openSerialCul(addr)
}

func openSerialCul(portName string) CulConn {
	mode := &serial.Mode{
		BaudRate: config.BaudRate,
		DataBits: 8,
		Parity:   serial.NoParity,
		StopBits: serial.OneStopBit,
	}
	port, err := serial.Open(portName, mode)
	if err != nil {
		slog.Error("Failed to open serial port", "port", portName, "error", err)
		return nil
	}
	slog.Info("Opened serial port", "port", portName, "baud", config.BaudRate)
	return &serialCulConn{port: port}
}

func openTCPCul(rawURL string) CulConn {
	hostPort, err := parseCulNetworkAddr(rawURL)
	if err != nil {
		slog.Error("Invalid CUL network URL", "url", rawURL, "error", err)
		return nil
	}

	conn, err := net.DialTimeout("tcp", hostPort, culDialTimeout)
	if err != nil {
		slog.Error("Failed to connect to CUL over TCP", "addr", hostPort, "error", err)
		return nil
	}

	slog.Info("Opened CUL TCP connection", "addr", hostPort)
	return &tcpCulConn{conn: conn}
}

func parseCulNetworkAddr(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	if u.Hostname() == "" {
		return "", fmt.Errorf("missing host")
	}
	port := u.Port()
	if port == "" {
		return "", fmt.Errorf("missing port")
	}
	return net.JoinHostPort(u.Hostname(), port), nil
}
