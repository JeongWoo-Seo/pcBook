package serializer

import (
	"fmt"
	"os"

	"google.golang.org/protobuf/proto"
)

func WriteProtobufToBinaryFile(message proto.Message, fileName string) error {
	data, err := proto.Marshal(message)
	if err != nil {
		return fmt.Errorf("can not marshal proto message to binary: %w", err)
	}

	err = os.WriteFile(fileName, data, 0644)
	if err != nil {
		return fmt.Errorf("can not write binary data to file: %w", err)
	}

	return nil
}

func ReadProtobufFromBinaryFile(filePath string, message proto.Message) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("can not read file: %w", err)
	}

	err = proto.Unmarshal(data, message)
	if err != nil {
		return fmt.Errorf("can not unmarshal binary file to proto message: %w", err)
	}

	return nil
}

func WriteProtobufToJsonFile(message proto.Message, fileName string) error {
	data, err := ProtobufToJson(message)
	if err != nil {
		return fmt.Errorf("can not marshal proto to json: %w", err)
	}

	err = os.WriteFile(fileName, []byte(data), 0644)
	if err != nil {
		return fmt.Errorf("can not wrtie json data to file: %w", err)
	}

	return nil
}
