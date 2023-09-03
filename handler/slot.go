package handler

import (
	"github.com/Tnze/go-mc/nbt"
	"github.com/Tnze/go-mc/net/packet"
	"io"
)

type ChangedSlots map[int]*Slot

func (c ChangedSlots) WriteTo(w io.Writer) (n int64, err error) {
	n, err = packet.VarInt(len(c)).WriteTo(w)
	if err != nil {
		return
	}
	for i, v := range c {
		n1, err := packet.Short(i).WriteTo(w)
		if err != nil {
			return n + n1, err
		}
		n2, err := v.WriteTo(w)
		if err != nil {
			return n + n1 + n2, err
		}
		n += n1 + n2
	}
	return
}

type Slot struct {
	ID    packet.VarInt
	Count packet.Byte
	NBT   nbt.RawMessage
}

func (s *Slot) WriteTo(w io.Writer) (n int64, err error) {
	var present packet.Boolean = s != nil
	return packet.Tuple{
		present,
		packet.Opt{
			Has: present,
			Field: packet.Tuple{
				&s.ID, &s.Count, packet.NBT(&s.NBT),
			},
		},
	}.WriteTo(w)
}

func (s *Slot) ReadFrom(r io.Reader) (n int64, err error) {
	var present packet.Boolean
	return packet.Tuple{
		&present, packet.Opt{
			Has: &present,
			Field: packet.Tuple{
				&s.ID, &s.Count, packet.NBT(&s.NBT),
			},
		},
	}.ReadFrom(r)
}
