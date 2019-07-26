package openpgp

import (
	"errors"
	"io"
)

// UserID represents a user identity. Implements Bindable.
type UserID struct {
	ID        []byte
	EnableMDC bool
}

// Packet returns an OpenPGP packet encoding this identity.
func (u *UserID) Packet() []byte {
	packet := make([]byte, len(u.ID)+2)
	packet[0] = 0xc0 | 13       // packet header, User ID Packet (13)
	packet[1] = byte(len(u.ID)) // packet length
	copy(packet[2:], u.ID)
	return packet
}

// Load from OpenPGP input (Packet() output).
func (u *UserID) Load(r io.Reader) (err error) {
	packet, err := readPacket(r)
	if err != nil {
		return err
	}
	if packet[0] != 0xc0|13 {
		return errors.New("invalid input")
	}

	u.ID = packet[2:]
	return nil
}

func (u *UserID) SignType() byte {
	return 0x13
}

func (u *UserID) Subpackets() []Subpacket {
	subpackets := []Subpacket{
		// Key Flags subpacket (sign and certify)
		// This is necessary since some implementations (GitHub) treat
		// all flags as if they were zero if not present.
		{Type: 27, Data: []byte{0x03}},
		// Features subpacket
		// This bit tells senders to use a Message Detection Code (MDC)
		// packet when encrypting messages. Data encrypted with OpenPGP
		// is, by default, unauthenticated! MDC is a mostly-broken form
		// of authentication that will make GnuPG complain a bit less.
		{Type: 30, Data: []byte{0x01}},
	}
	if u.EnableMDC {
		return subpackets[:]
	} else {
		return subpackets[:1]
	}
}

func (u *UserID) SignData() []byte {
	prefix := []byte{0xb4, 0, 0, 0}
	packet := u.Packet()[1:]
	return append(prefix, packet...)
}