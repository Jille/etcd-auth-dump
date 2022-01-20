// A binary to dump etcd auth configuration as a list of shell commands suitable for an empty cluster.
package main

import (
	"context"
	"fmt"
	"log"

	authdump "github.com/Jille/etcd-auth-dump"
	clientconfig "github.com/Jille/etcd-client-from-env"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
)

func main() {
	ctx := context.Background()
	cc, err := clientconfig.Get()
	if err != nil {
		log.Fatalf("Failed to parse environment settings: %v", err)
	}
	cc.DialOptions = append(cc.DialOptions, grpc.WithBlock())
	c, err := clientv3.New(cc)
	if err != nil {
		log.Fatalf("Failed to connect to etcd: %v", err)
	}
	defer c.Close()

	cmds, _, err := authdump.Dump(ctx, c, 0)
	if err != nil {
		log.Fatal(err)
	}
	for _, c := range cmds {
		fmt.Println(c)
	}
}
