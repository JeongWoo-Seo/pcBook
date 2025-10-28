package serializer

import (
	"fmt"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func ProtobufToJson(message proto.Message) (string, error) {
	marshaler := protojson.MarshalOptions{
		Indent:         "  ",
		UseProtoNames:  true,
		UseEnumNumbers: true,
	}

	data, err := marshaler.Marshal(message)
	if err != nil {
		return "", fmt.Errorf("could not marshal proto message to JSON: %w", err)
	}

	return string(data), nil
}
