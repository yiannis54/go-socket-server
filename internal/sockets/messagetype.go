package sockets

import pb "github.com/yiannis54/go-socket-server/notificationspb" // Replace with actual import path

// MessageType is the type of notification messages.
type MessageType string

// Notification message types.
const (
	TypeError MessageType = "error"
	TypeInfo  MessageType = "info"
)

// ToProtoEnum converts your string MessageType to the protobuf enum
func (mt MessageType) ToProtoEnum() pb.MessageType {
	switch mt {
	case TypeError:
		return pb.MessageType_TYPE_ERROR
	case TypeInfo:
		return pb.MessageType_TYPE_INFO
	default:
		return pb.MessageType_MESSAGE_TYPE_UNSPECIFIED
	}
}

// FromProtoEnum converts the protobuf enum to your string MessageType
func FromProtoEnum(mt pb.MessageType) MessageType {
	switch mt {
	case pb.MessageType_TYPE_ERROR:
		return TypeError
	case pb.MessageType_TYPE_INFO:
		return TypeInfo
	default:
		return MessageType("") // or return a default like "unknown"
	}
}
