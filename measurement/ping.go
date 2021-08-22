package measurement

// Represents a scamper "ping" task
//
// Implements ScCommand
type Ping struct {
	TCPAck     uint32 `short:"A" help:"Number to use in the acknowledgement field of the TCP header, or the sequence number field of the TCP header when sending reset probes"`
	Payload    string `short:"B" help:"Payload to include in each probe (hexadecimal string)"`
	ProbeCount uint16 `short:"c" default:"4" help:"Number of probes to send before exiting"`
	ICMPSum    uint16 `short:"C" help:"ICMP checksum to use when sending a probe. The payload of each probe will be manipulated so that the checksum is valid."`
	DstPort    uint16 `short:"d" help:"Destination port to use in each TCP/UDP probe, and the first ICMP sequence number to use in ICMP probes."`
	SrcPort    uint16 `short:"F" help:"Source port to use in each TCP/UDP probe, and the ICMP ID to use in ICMP probes."`
	Wait       uint8  `short:"i" default:"1" help:"Length of time to wait, in seconds, between probes."`
	TTL        uint8  `short:"m" default:"64" help:"TTL value to use for outgoing packets."`
	MTU        uint16 `short:"M" help:"Pseudo MTU value. If the response packet is larger than the pseudo MTU, an ICMP packet too big (PTB) message is sent."`
	ReplyCount uint16 `short:"o" help:"Number of replies required at which time probing may cease. By default, all probes are sent"`
	// TODO: -O options
	Pattern     string     `short:"p" help:"Pattern, in hex, to use in probes. Up to 16 bytes may be specified. By default, each probe’s bytes are zeroed."`
	Method      PingMethod `short:"P" default:"icmp-echo" help:"Type of ping packets to send."`
	RouterAddr  string     `short:"r" help:"IP address of the router to use."`
	RecordRoute bool       `short:"R" help:"Specifies that the record route IP option should be used."`
	Size        uint16     `short:"s" help:"Size of the probes to send. The probe size includes the length of the IP and ICMP headers. By default, a probe size of 84 bytes is used for IPv4 pings, and 56 bytes for IPv6 pings."`
	SrcAddr     string     `short:"S" help:"Source address to use in probes. The address can be spoofed if Options.Spoof is set."`
	Timestamp   string     `short:"T" help:"Specifies that an IP timestamp option be included."`
	Timeout     uint8      `short:"W" default:"1" help:"How long to wait for responses after the last ping is sent."`
}

//go:generate enumer -type=PingMethod -json -text -linecomment
type PingMethod uint8

const (
	ICMP_ECHO     PingMethod = iota // icmp-echo
	ICMP_TIME                       // icmp-time
	TCP_SYN                         // tcp-syn
	TCP_ACK                         // tcp-ack
	TCP_ACK_SPORT                   // tcp-ack-sport
	TCP_SYNACK                      // tcp-synack
	TCP_RST                         // tcp-rst
	UDP                             // udp
	UDP_DPORT                       // udp-dport
)

func (p Ping) AsCommand() string {
	// TODO
	return ""
}
