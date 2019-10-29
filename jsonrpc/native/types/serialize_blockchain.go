package types

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
)

func check0xPrefix(s string) error {
	if !strings.HasPrefix(s, "0x") {
		return fmt.Errorf("invalid value, should be 0x-prefix")
	}
	return nil
}

// Serialize hash
func (h *Hash) Serialize() ([]byte, error) {
	inner := string(*h)

	err := check0xPrefix(inner)
	if err != nil {
		return nil, err
	}

	b, err := hex.DecodeString(inner[2:])
	if err != nil {
		return nil, err
	}

	if len(b) != 32 {
		return nil, fmt.Errorf("invalid hash, should be 32 bytes")
	}

	return b, nil
}

// Serialize script hash type
func (t *ScriptHashType) Serialize() ([]byte, error) {
	inner := string(*t)

	if strings.Compare(inner, string(Data)) != 0 && strings.Compare(inner, string(Type)) != 0 {
		return nil, fmt.Errorf("invalid script hash type")
	}

	if strings.Compare(inner, string(Data)) == 0 {
		return []byte{00}, nil
	}

	return []byte{01}, nil
}

// Serialize dep type
func (t *DepType) Serialize() ([]byte, error) {
	inner := string(*t)

	if strings.Compare(inner, string(Code)) != 0 && strings.Compare(inner, string(DepGroup)) != 0 {
		return nil, fmt.Errorf("invalid dep group")
	}

	if strings.Compare(inner, string(Code)) == 0 {
		return []byte{00}, nil
	}

	return []byte{01}, nil
}

// Serialize bytes
func (b *Bytes) Serialize() ([]byte, error) {
	inner := string(*b)

	err := check0xPrefix(inner)
	if err != nil {
		return nil, err
	}

	// Fixvec, vector Bytes <byte>
	if len(inner[2:]) == 0 {
		return []byte{00, 00, 00, 00}, nil
	}

	bs, err := hex.DecodeString(inner[2:])
	if err != nil {
		return nil, err
	}

	return bs, nil
}

// Serialize uint32
func (u *Uint32) Serialize() ([]byte, error) {
	inner := string(*u)

	err := check0xPrefix(inner)
	if err != nil {
		return nil, err
	}

	uu := inner[2:]
	if len(inner)%2 != 0 {
		uu = "0" + uu
	}

	n, err := strconv.ParseUint(uu, 16, 32)
	if err != nil {
		return nil, err
	}

	return serializeUint32(uint32(n)), nil
}

// Serialize uint64
func (u *Uint64) Serialize() ([]byte, error) {
	inner := string(*u)

	err := check0xPrefix(inner)
	if err != nil {
		return nil, err
	}

	uu := inner[2:]
	if len(inner)%2 != 0 {
		uu = "0" + uu
	}

	n, err := strconv.ParseUint(uu, 16, 64)
	if err != nil {
		return nil, err
	}

	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, n)

	return b, nil
}

// Serialize script
func (s *Script) Serialize() ([]byte, error) {
	h, err := s.CodeHash.Serialize()
	if err != nil {
		return nil, err
	}

	t, err := s.HashType.Serialize()
	if err != nil {
		return nil, err
	}

	a, err := s.Args.Serialize()
	if err != nil {
		return nil, err
	}

	size := 4 + 4*3 + len(h) + len(t) + len(a) + 4
	hOffset := 4 + 4*3
	tOffset := hOffset + len(h)
	aOffset := tOffset + len(t)

	b := new(bytes.Buffer)

	b.Write(serializeUint32(uint32(size)))
	b.Write(serializeUint32(uint32(hOffset)))
	b.Write(serializeUint32(uint32(tOffset)))
	b.Write(serializeUint32(uint32(aOffset)))
	b.Write(h)
	b.Write(t)
	b.Write(serializeUint32(uint32(len(a))))
	b.Write(a)

	return b.Bytes(), nil
}

// Serialize outpoint
func (o *OutPoint) Serialize() ([]byte, error) {
	h, err := o.TxHash.Serialize()
	if err != nil {
		return nil, err
	}

	i, err := o.Index.Serialize()
	if err != nil {
		return nil, err
	}

	b := new(bytes.Buffer)

	b.Write(h)
	b.Write(i)

	return b.Bytes(), nil
}

// Serialize cell input
func (i *CellInput) Serialize() ([]byte, error) {
	s, err := i.Since.Serialize()
	if err != nil {
		return nil, err
	}

	o, err := i.PreviousOutput.Serialize()
	if err != nil {
		return nil, err
	}

	b := new(bytes.Buffer)

	b.Write(s)
	b.Write(o)

	return b.Bytes(), nil
}

// Serialize cell output
func (o *CellOutput) Serialize() ([]byte, error) {
	c, err := o.Capacity.Serialize()
	if err != nil {
		return nil, err
	}

	l, err := o.Lock.Serialize()
	if err != nil {
		return nil, err
	}

	var t []byte
	if o.Type != nil {
		t, err = o.Type.Serialize()
		if err != nil {
			return nil, err
		}
	}

	size := 4 + 4*3 + len(c) + len(l)
	if len(t) != 0 {
		size += len(t)
	}
	cOffset := 4 + 4*3
	lOffset := cOffset + len(c)
	tOffset := lOffset + len(l)

	b := new(bytes.Buffer)

	b.Write(serializeUint32(uint32(size)))
	b.Write(serializeUint32(uint32(cOffset)))
	b.Write(serializeUint32(uint32(lOffset)))
	b.Write(serializeUint32(uint32(tOffset)))
	b.Write(c)
	b.Write(l)
	b.Write(t)

	return b.Bytes(), nil
}

// Serialize cell dep
func (d *CellDep) Serialize() ([]byte, error) {
	o, err := d.OutPoint.Serialize()
	if err != nil {
		return nil, err
	}

	dd, err := d.DepType.Serialize()
	if err != nil {
		return nil, err
	}

	b := new(bytes.Buffer)

	b.Write(o)
	b.Write(dd)

	return b.Bytes(), nil
}

// Serialize transaction
func (t *Transaction) Serialize() ([]byte, error) {
	v, err := t.Version.Serialize()
	if err != nil {
		return nil, err
	}

	cds := make([][]byte, len(t.CellDeps))
	for i := 0; i < len(t.CellDeps); i++ {
		cd, err := t.CellDeps[i].Serialize()
		if err != nil {
			return nil, err
		}

		cds[i] = cd
	}
	cdsBytes := SerializeFixVec(cds)

	hds := make([][]byte, len(t.HeaderDeps))
	for i := 0; i < len(t.HeaderDeps); i++ {
		hd, err := t.HeaderDeps[i].Serialize()
		if err != nil {
			return nil, err
		}

		hds[i] = hd
	}
	hdsBytes := SerializeFixVec(hds)

	ips := make([][]byte, len(t.Inputs))
	for i := 0; i < len(t.Inputs); i++ {
		ip, err := t.Inputs[i].Serialize()
		if err != nil {
			return nil, err
		}

		ips[i] = ip
	}
	ipsBytes := SerializeFixVec(ips)

	ops := make([][]byte, len(t.Outputs))
	for i := 0; i < len(t.Outputs); i++ {
		op, err := t.Outputs[i].Serialize()
		if err != nil {
			return nil, err
		}

		ops[i] = op
	}
	opsBytes := SerializeDynVec(ops)

	ods := make([][]byte, len(t.OutputsData))
	for i := 0; i < len(t.OutputsData); i++ {
		od, err := t.OutputsData[i].Serialize()
		if err != nil {
			return nil, err
		}

		ods[i] = od
	}
	odsBytes := SerializeDynVec(ods)

	fields := [][]byte{v, cdsBytes, hdsBytes, ipsBytes, opsBytes, odsBytes}
	return SerializeTable(fields), nil
}
