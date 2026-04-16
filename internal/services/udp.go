package services

import "fmt"

type UDPDecoyService struct{}

func NewUDPDecoyService() *UDPDecoyService {
	return &UDPDecoyService{}
}

func (s *UDPDecoyService) Name() string {
	return "udp-decoy"
}

func (s *UDPDecoyService) HandleConn(*SessionContext) error {
	return nil
}

func (s *UDPDecoyService) HandlePacket(ctx *PacketContext) error {
	_ = ctx.Recorder.Event(ctx.Context, ctx.Session, "udp.datagram", map[string]any{
		"size":            len(ctx.Payload),
		"payload_preview": string(ctx.Payload),
	})
	return ctx.Write([]byte(fmt.Sprintf("%s: request accepted", s.Name())))
}
