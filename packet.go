package main

import(
	"fmt"
	"encoding/binary"
)

type Packet struct{
	srcIdentity [4]byte
	dstIdentity [4]byte
	Length uint64
	Payload []byte
}

func DecodePacket(pktbytes []byte, pkt *Packet)(error){

	if len(pktbytes)<16{
		return fmt.Errorf("Invalid packet received. Incomplete header")
	}
	if len(pktbytes)==16{
		return fmt.Errorf("Empty packet body")
	}

	copy(pkt.srcIdentity[:], pktbytes[:4])
	copy(pkt.dstIdentity[:], pktbytes[4:8])
	pkt.Length = binary.BigEndian.Uint64(pktbytes[8:16])
	pkt.Payload = append(pkt.Payload, pktbytes[16:]...)

	return nil
}

func (p Packet)Encode()([]byte, error){
//      starttime := time.Now()
        buf:=make([]byte, 16)           // 4 + 4 + 8(uint64) = 16

        copy(buf[:4], p.srcIdentity[:])
//      fmt.Println("Copied in srcidentity=", buf)
        copy(buf[4:8], p.dstIdentity[:])
//      fmt.Println("copied in destidentity=", buf)
        binary.BigEndian.PutUint64(buf[8:16], uint64(p.Length))
//      fmt.Println("copied in length=", buf)
        buf = append(buf, p.Payload...)
//      fmt.Println("copied in payload=", buf)

//      endtime:= time.Since(starttime)
//      fmt.Println("Time taken in packet encoder=", endtime)

        // append an io.EOF to the end of the packet bytes
        buf = append(buf, byte('\x00'))

        return buf, nil
}
