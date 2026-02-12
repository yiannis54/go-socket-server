package app

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"

	"github.com/yiannis54/go-socket-server/internal/config"
	"github.com/yiannis54/go-socket-server/internal/notifications"
	"github.com/yiannis54/go-socket-server/internal/sockets"
	pb "github.com/yiannis54/go-socket-server/notificationspb"
)

type NotificationServer struct {
	pb.UnimplementedNotificationServiceServer
	notificationsClient *notifications.Client
}

func runRpc(ctx context.Context, notificationsClient *notifications.Client, cfg *config.EnvConfig) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
	if err != nil {
		return fmt.Errorf("failed to listen rpc: %w", err)
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(authInterceptor),
	)
	pb.RegisterNotificationServiceServer(grpcServer, &NotificationServer{
		notificationsClient: notificationsClient,
	})

	done := make(chan struct{})
	go func() {
		log.Printf("Listening RPC server on :%v\n", cfg.GRPCPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Printf("failed to serve rpc: %v\n", err)
		}
		close(done)
	}()

	select {
	case <-ctx.Done():
		log.Println("Received shutdown signal for RPC server")
		grpcServer.GracefulStop()
		<-done // Wait for Serve to return
		return nil
	case <-done:
		// Server exited unexpectedly
		return fmt.Errorf("gRPC server exited unexpectedly")
	}
}

func (s *NotificationServer) Broadcast(ctx context.Context, msg *pb.Message) (*empty.Empty, error) {
	s.notificationsClient.Broadcast(ctx, &sockets.Message{
		Type:        sockets.FromProtoEnum(msg.Type),
		EntityID:    msg.EntityId,
		MessageBody: msg.Message,
	})
	return &empty.Empty{}, nil
}

func (s *NotificationServer) NotifyRoom(ctx context.Context, msg *pb.MessageWithRoom) (*empty.Empty, error) {
	s.notificationsClient.NotifyRoom(ctx, &sockets.MessageWithRoom{
		Message: sockets.Message{
			Type:        sockets.FromProtoEnum(msg.Base.Type),
			EntityID:    msg.Base.EntityId,
			MessageBody: msg.Base.Message,
		},
		RoomName: msg.Room,
	})
	return &empty.Empty{}, nil
}

func (s *NotificationServer) PrivateNotify(ctx context.Context, msg *pb.MessageWithUser) (*empty.Empty, error) {
	s.notificationsClient.PrivateNotify(ctx, &sockets.MessageWithUser{
		Message: sockets.Message{
			Type:        sockets.FromProtoEnum(msg.Base.Type),
			EntityID:    msg.Base.EntityId,
			MessageBody: msg.Base.Message,
		},
		UserID: msg.UserId,
	})
	return &empty.Empty{}, nil
}

func authInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	// Implement gRPC authentication.
	// return nil, errors.New("failed authentication")

	return handler(ctx, req)
}
