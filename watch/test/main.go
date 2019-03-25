package main

import (
	"fmt"
	"github.com/coreos/etcd/etcdserver/etcdserverpb"
	"github.com/pingcap/pd/watch"
	"google.golang.org/grpc"
	"math"
	"log"
	"context"
)

func main(){
	conn, err := grpc.Dial("localhost:28080", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("faild to connect: %v", err)
	} else {
		fmt.Println("connection success")
	}
	defer conn.Close()

	watchClient := etcdserverpb.NewWatchClient(conn)

	watcher := watch.NewWatcherWithCallOption(watchClient,[]grpc.CallOption{
		grpc.FailFast(false),
		grpc.MaxCallSendMsgSize(2 * 1024 * 1024),
		grpc.MaxCallRecvMsgSize(math.MaxInt32),
	})



	for{
		rch := watcher.Watch(context.Background(), "mykey")
		for wresp := range rch {
			fmt.Println(wresp.Header)
			fmt.Println("we have receive new event")
		}
	}
}
