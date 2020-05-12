package rpc

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
)

// WatchHealth calls health/v1 Watch method a invokes the passed
// callback whenever a service status changes
func WatchHealth(address string, callback func(alive bool, err error)) {
	for {
		conn, err := grpc.Dial(address, grpc.WithInsecure())
		if err != nil {
			callback(false, err)
			return
		}

		ctx := context.Background()
		defer conn.Close()

		ticker := time.NewTicker(4 * time.Second)
		defer ticker.Stop()

	loop:
		for err == nil {
			select {
			case <-ticker.C:
				resp, err := grpc_health_v1.NewHealthClient(conn).Check(ctx, &grpc_health_v1.HealthCheckRequest{Service: "anycable.RPC"})
				if err != nil {
					callback(false, err)
					break loop
				}

				status := resp.GetStatus()

				if status == grpc_health_v1.HealthCheckResponse_SERVICE_UNKNOWN {
					log.Fatal("gRPC health check failed with service unknown error")
				} else {
					callback(status == grpc_health_v1.HealthCheckResponse_SERVING, nil)
				}
			}
		}

		time.Sleep(5 * time.Second)
	}
}
