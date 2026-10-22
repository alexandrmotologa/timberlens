package ingest

import (
	"net"
	"sync/atomic"

	"github.com/alexandrmotologa/timberlens/internal/parser"
)

// SyslogListener listens on a UDP port for RFC 3164/5424 syslog packets.
type SyslogListener struct {
	addr     string
	consumer LineConsumer
	seqID    uint64
	running  int32
	conn     *net.UDPConn
}

// NewSyslogListener creates a syslog receiver.
func NewSyslogListener(addr string, consumer LineConsumer) *SyslogListener {
	if addr == "" {
		addr = ":1514"
	}
	return &SyslogListener{
		addr:     addr,
		consumer: consumer,
	}
}

// Start opens the UDP socket and reads syslog datagrams.
func (sl *SyslogListener) Start() error {
	laddr, err := net.ResolveUDPAddr("udp", sl.addr)
	if err != nil {
		return err
	}

	conn, err := net.ListenUDP("udp", laddr)
	if err != nil {
		return err
	}
	sl.conn = conn

	atomic.StoreInt32(&sl.running, 1)
	defer atomic.StoreInt32(&sl.running, 0)

	buf := make([]byte, 65535)

	for atomic.LoadInt32(&sl.running) == 1 {
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			if atomic.LoadInt32(&sl.running) == 0 {
				return nil
			}
			continue
		}

		if n == 0 {
			continue
		}

		packet := make([]byte, n)
		copy(packet, buf[:n])

		seq := atomic.AddUint64(&sl.seqID, 1)
		rec, _ := parser.ParseLine(packet, seq)
		if rec != nil && sl.consumer != nil {
			sl.consumer(rec)
		}
	}

	return nil
}

// Stop closes the UDP socket.
func (sl *SyslogListener) Stop() {
	atomic.StoreInt32(&sl.running, 0)
	if sl.conn != nil {
		sl.conn.Close()
	}
}
