package icmp

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/sewh/cping/config"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
)

const (
	IPv4Len       = 20
	ICMPv4PingLen = 8

	IPv6Len       = 40
	ICMPv6PingLen = 8
)

var (
	// icmp errors
	TimeoutExceeded  = errors.New("timeout exceeded")
	TTLExpired       = errors.New("ttl exceeded")
	SourceQuench     = errors.New("source quench")
	DestUnreachable  = errors.New("destination unreachable")
	CouldNotFragment = errors.New("could not fragment")
	UnknownPacket    = errors.New("unknown response packet")

	// general errors
	BadIPVersion = errors.New("bad ip protocol version")
	BadIPAddress = errors.New("malformed ip address")
)

type Result struct {
	Error    error
	Sent     time.Time
	Received time.Time
}

type Sender struct {
	Config   *config.Config
	ID       uint16
	Seq      uint16
	Conn     *net.IPConn
	IPv4Conn *ipv4.Conn
	IPv6Conn *ipv6.Conn
	Results  []Result
}

func NewSender(c *config.Config) *Sender {
	return &Sender{
		Config: c,
		ID:     0,
		Seq:    0,
		Conn:   nil,
	}
}

func PercentOf(part int, total int) float64 {
	return (float64(part) * float64(100)) / float64(total)
}

func (s *Sender) Stats() string {
	attempts := 0
	succeeded := 0
	minrtt := 0
	maxrtt := 0
	avgrtt := 0
	var percent float64 = 0.0

	for _, res := range s.Results {
		attempts += 1
		if res.Error == nil {
			succeeded += 1
		} else {
			continue
		}

		rtt := int(res.Received.UnixMilli() - res.Sent.UnixMilli())
		if rtt < minrtt || minrtt == 0 {
			minrtt = rtt
		}
		if rtt > maxrtt {
			maxrtt = rtt
		}
		avgrtt += rtt
	}

	if succeeded > 0 {
		avgrtt /= succeeded
		percent = PercentOf(succeeded, attempts)
	}

	return fmt.Sprintf("Success rate is %.1f percent (%d/%d), round-trip min/avg/max = %d/%d/%d ms",
		percent, succeeded, attempts, minrtt, avgrtt, maxrtt)

}

func (s *Sender) SendAndReceive() error {
	// make sure we have a ping ID generated
	s.EnsureID()

	// make sure we have an open raw socket
	err := s.EnsureSocketOpen()
	if err != nil {
		return err
	}

	// craft ping packet
	pkt, err := s.CraftPacket()
	if err != nil {
		return err
	}

	// send packet
	to, err := net.ResolveIPAddr(fmt.Sprintf("ip%d", s.Config.IPVersion), s.Config.DestIP)
	if err != nil {
		return BadIPAddress
	}
	_, err = s.Conn.WriteTo(pkt, to)
	if err != nil {
		return err
	}
	defer func() { s.Seq += 1 }()

	r := &Result{
		Sent: time.Now(),
	}

	// receive a response
	res := s.Receive()

	// update stats
	r.Received = time.Now()
	r.Error = res

	s.Results = append(s.Results, *r)

	return res
}

func (s *Sender) CraftPacket() ([]byte, error) {
	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}

	switch s.Config.IPVersion {
	case 4:
		payloadLen := s.Config.Size - (IPv4Len + ICMPv4PingLen)
		if payloadLen < 0 {
			payloadLen = 0
		}

		gopacket.SerializeLayers(buf, opts,
			&layers.ICMPv4{
				TypeCode: layers.CreateICMPv4TypeCode(8, 0),
				Id:       s.ID,
				Seq:      s.Seq,
			},
			gopacket.Payload(s.CyclePayload(payloadLen)),
		)
	case 6:
		payloadLen := s.Config.Size - (IPv6Len + ICMPv6PingLen)
		if payloadLen < 0 {
			payloadLen = 0
		}

		gopacket.SerializeLayers(buf, opts,
			&layers.ICMPv6{
				TypeCode: layers.CreateICMPv6TypeCode(128, 0),
			},
			&layers.ICMPv6Echo{
				Identifier: s.ID,
				SeqNumber:  s.Seq,
			},
			gopacket.Payload(s.CyclePayload(payloadLen)),
		)

	default:
		return nil, BadIPVersion
	}

	return buf.Bytes(), nil
}

func (s *Sender) Receive() error {
	var err error
	buf := make([]byte, s.Config.Size)

	// set a deadline on the socket
	deadline := time.Now().Add(
		time.Second * time.Duration(s.Config.TimeoutSecs),
	)
	err = s.Conn.SetReadDeadline(deadline)
	if err != nil {
		return err
	}

	for {
		amt, err := s.Conn.Read(buf)

		// make sure we haven't gone past our timeout
		if errors.Is(err, os.ErrDeadlineExceeded) {
			return TimeoutExceeded
		} else if err != nil {
			return err
		}

		// check if the packet is related to us
		switch s.Config.IPVersion {
		case 4:

			// parse packet as ICMPv4
			pkt := gopacket.NewPacket(
				buf[:amt], layers.LayerTypeIPv4, gopacket.Default,
			)

			if icmpLayer := pkt.Layer(layers.LayerTypeICMPv4); icmpLayer != nil {
				icmp, _ := icmpLayer.(*layers.ICMPv4)

				// check that the ID is something we expect
				if icmp.Id != s.ID {
					continue
				}

				// return result
				t := icmp.TypeCode.Type()
				c := icmp.TypeCode.Code()

				switch t {

				case layers.ICMPv4TypeEchoReply:
					return nil

				case layers.ICMPv4TypeDestinationUnreachable:

					if c == layers.ICMPv4CodeFragmentationNeeded {
						return CouldNotFragment
					}

					return DestUnreachable
				case layers.ICMPv4TypeSourceQuench:
					return SourceQuench

				case layers.ICMPv4TypeTimeExceeded:
					return TTLExpired

				default:
					return UnknownPacket
				}

			} else {
				continue
			}

		case 6:

			// parse packet as ICMPv6
			pkt := gopacket.NewPacket(
				buf[:amt], layers.LayerTypeICMPv6, gopacket.Default,
			)

			if icmpLayer := pkt.Layer(layers.LayerTypeICMPv6); icmpLayer != nil {
				icmp, _ := icmpLayer.(*layers.ICMPv6)

				t := icmp.TypeCode.Type()
				c := icmp.TypeCode.Code()
				c = c

				switch t {
				case layers.ICMPv6TypeEchoReply:

					if echoLayer := pkt.Layer(layers.LayerTypeICMPv6Echo); echoLayer != nil {
						echo, _ := echoLayer.(*layers.ICMPv6Echo)

						if echo.Identifier == s.ID {
							return nil
						}
					} else {
						return UnknownPacket
					}
				case layers.ICMPv6TypeDestinationUnreachable:
					return DestUnreachable
				case layers.ICMPv6TypePacketTooBig:
					return CouldNotFragment
				case layers.ICMPv6TypeTimeExceeded:
					return TTLExpired
				default:
					return UnknownPacket
				}
			}
		}
	}

	return nil
}

func (s *Sender) CyclePayload(amt int) []byte {
	pos := 0
	out := make([]byte, amt)

	for i := 0; i < amt; i += 1 {
		if pos == len(s.Config.Payload)-1 {
			pos = 0
		} else {
			pos += 1
		}
		out[i] = s.Config.Payload[pos]
	}

	return out
}

func (s *Sender) EnsureID() error {
	if s.ID != 0 {
		return nil
	}

	buf := make([]byte, 2)

	_, err := rand.Read(buf)
	if err != nil {
		return err
	}

	s.ID = binary.BigEndian.Uint16(buf)

	return nil
}

func (s *Sender) EnsureSocketOpen() error {
	if s.Conn != nil {
		return nil
	}

	i := "icmp"
	switch s.Config.IPVersion {
	case 4:
		i = "icmp"
	case 6:
		i = "icmp6"
	default:
		return BadIPVersion
	}

	conn, err := net.ListenIP(fmt.Sprintf("ip%d:%s", s.Config.IPVersion, i), nil)
	if err != nil {
		return err
	}

	switch s.Config.IPVersion {
	case 4:
		s.IPv4Conn = ipv4.NewConn(conn)
		s.IPv4Conn.SetTTL(s.Config.TTL)
	case 6:
		s.IPv6Conn = ipv6.NewConn(conn)
		s.IPv6Conn.SetHopLimit(s.Config.TTL)
	default:
		return BadIPVersion
	}

	s.Conn = conn

	return nil
}

func (s *Sender) Close() error {
	return s.Conn.Close()
}
