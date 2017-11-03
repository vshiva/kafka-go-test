package types

import "github.com/golang/protobuf/proto"

// ProtobufEncoder for Kafka Message
type ProtobufEncoder struct {
	Msg *TasProtocol
}

//Encode the contained proto message
func (p ProtobufEncoder) Encode() ([]byte, error) {
	return proto.Marshal(p.Msg)
}

//Length of the encode messge
func (p ProtobufEncoder) Length() int {
	return proto.Size(p.Msg)
}
