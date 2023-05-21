package broadcastdispatcher

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"sync"
	"time"
	shootoutpb "wildwest/api/proto/shootout"
)

// BroadcastShootoutTime waits until all cowboys are ready and broadcasts to them when to begin the shootout
func BroadcastShootoutTime(logger *zap.Logger, replicas int, podName string, serviceName string, grpcPort int) {
	// wait until all replicas are ready by looking up hostname
	var wg sync.WaitGroup
	wg.Add(replicas)

	for i := 0; i < replicas; i++ {
		go func(wg *sync.WaitGroup, id int) {
			defer wg.Done()

			// build hostname
			hostname := fmt.Sprintf("%s-%d.%s", podName, id, serviceName)

			retryCount := 0

			// TODO exit after N retries, use exponential backoff
			_, err := net.LookupHost(hostname)
			for err != nil {
				retryCount++
				if retryCount > 20 {
					logger.Warn("retry count above 20", zap.String("hostname", hostname))
				}

				time.Sleep(5 * time.Second)

				_, err = net.LookupHost(hostname)
			}
		}(&wg, i)
	}
	wg.Wait()

	// begin shootout in 10 seconds from now
	shootoutTime := time.Now().Add(10 * time.Second).Round(time.Second)
	shootoutTimestamp := shootoutTime.Unix()

	logger.Info("broadcasting shootout beginning time...", zap.Time("shootout_time", shootoutTime))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(replicas/5+2)*time.Second)
	defer cancel()

	wg.Add(replicas)

	// broadcast the shootout time
	for i := 0; i < replicas; i++ {
		go func(ctx context.Context, id int) {
			defer wg.Done()

			// build hostname
			hostname := fmt.Sprintf("%s-%d.%s:%d", podName, id, serviceName, grpcPort)

			// dial cowboy
			conn, err := grpc.Dial(hostname, grpc.WithTransportCredentials(insecure.NewCredentials())) // TODO insecure
			if err != nil {
				logger.Fatal("failed to dial", zap.Error(err))
			}

			// create client
			client := shootoutpb.NewShootoutServiceClient(conn)

			// begin shootout
			if _, err := client.ReceiveShootoutTime(ctx, &shootoutpb.ReceiveShootoutTimeRequest{Timestamp: shootoutTimestamp}); err != nil {
				logger.Error("failed to send shootout beginning time", zap.Error(err))
			}

			conn.Close()
		}(ctx, i)
	}

	wg.Wait()
}
