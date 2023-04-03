package broadcastdispatcher

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	shootoutpb "wildwest/api/proto/shootout"
)

// BroadcastShootoutTime waits until all cowboys are ready and broadcasts to them when to begin the shootout
func BroadcastShootoutTime(logger *zap.Logger, replicas int, podName string, serviceName string, namespace string, grpcPort string) {
	// wait until all replicas are ready by looking up hostname
	var wg sync.WaitGroup
	for i := 0; i < replicas; i++ {
		wg.Add(1)
		go func(wg *sync.WaitGroup, id int) {
			defer wg.Done()
			// build hostname
			hostname := fmt.Sprintf("%s-%d.%s.%s.svc.cluster.local", podName, id, serviceName, namespace)

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

	logger.Info("broadcasting beginning time of shootout...", zap.Time("shootout_time", shootoutTime))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(replicas/5+2)*time.Second)
	defer cancel()

	// broadcast the shootout time
	for i := 0; i < replicas; i++ {
		wg.Add(1)
		go func(ctx context.Context, id int) {
			// build hostname
			hostname := fmt.Sprintf("%s-%d.%s.%s.svc.cluster.local%s", podName, id, serviceName, namespace, grpcPort)

			// dial cowboy
			conn, err := grpc.Dial(hostname, grpc.WithTransportCredentials(insecure.NewCredentials())) // TODO insecure
			if err != nil {
				logger.Fatal("failed to dial", zap.Error(err))
			}

			// create client
			client := shootoutpb.NewShootoutServiceClient(conn)

			// begin shootout
			if _, err := client.BeginShootout(ctx, &shootoutpb.BeginShootoutRequest{Timestamp: shootoutTimestamp}); err != nil {
				logger.Error("failed to begin shootout", zap.Error(err))
			}

			conn.Close()
		}(ctx, i)
	}

	wg.Wait()
}
