package dbinit

import (
	"context"
	"log"
	"net"
	"net/http"

	"github.com/gorilla/mux"
	"gitlab.com/Sh00ty/hootydb/internal/config"
	"gitlab.com/Sh00ty/hootydb/internal/kv/memmap"
	"gitlab.com/Sh00ty/hootydb/internal/replicator"
	"gitlab.com/Sh00ty/hootydb/internal/replicator/quorum"
	"gitlab.com/Sh00ty/hootydb/internal/sequence/inmemseq"
	"gitlab.com/Sh00ty/hootydb/internal/server/localsrv"
	"gitlab.com/Sh00ty/hootydb/internal/server/public/httpsrv"
	"gitlab.com/Sh00ty/hootydb/internal/transport/clients/localclnt"
	"gitlab.com/Sh00ty/hootydb/internal/transport/clients/remote"
	"gitlab.com/Sh00ty/hootydb/internal/transport/transport"
	"gitlab.com/Sh00ty/hootydb/internal/utils/logger"
	"google.golang.org/grpc"
)

// label is almost same as dbName, but its need for logging labeling
func StartDb(ctx context.Context, env string, dbName string, dbLabel string) {
	var (
		nodeCfg = config.GetNodeCfg(dbName)
		cfg     = config.NewCfg(2, 2)
		kv      = memmap.New(10)
		clients = make(map[string]replicator.Client, len(nodeCfg.RemoteAddrs))
	)

	for _, remoteAddr := range nodeCfg.RemoteAddrs {
		clnt, err := remote.NewReplicatorClient(remoteAddr)
		if err != nil {
			log.Panicf("failed to create replication client for addr %v: %v", remoteAddr, err)
		}
		clients[remoteAddr] = clnt
	}
	localClnt := localclnt.New(kv, nodeCfg.DBFaultPercentage)
	clients[nodeCfg.SelfAddr] = localClnt
	seq := inmemseq.NewInMemSeq()

	log := logger.NewLogger(env, dbLabel)
	replicator := quorum.NewReplicator(cfg, seq, clients, log)
	localSrv := localsrv.New(kv, nodeCfg.RpcFaultPercentage)
	publicSrv := httpsrv.NewSrv(replicator, kv, log)
	go func() {
		router := mux.NewRouter()
		router.HandleFunc("/db/get/{key}/", publicSrv.Read).
			Methods("GET").
			Schemes("http")
		router.HandleFunc("/db/set/", publicSrv.Write).
			Methods("POST").
			Schemes("http")

		router.HandleFunc("/db/stale/get/{key}/", publicSrv.StaleRead).
			Methods("GET").
			Schemes("http")
		router.HandleFunc("/db/stale/set/", publicSrv.StaleWrite).
			Methods("POST").
			Schemes("http")

		log.Infof(ctx, "public srv %s started on %s", dbName, nodeCfg.PublicAddr)
		err := http.ListenAndServe(nodeCfg.PublicAddr, router)
		if err != nil {
			log.Fatalf(ctx, "error in public http server start on %s", dbName)
		}
	}()

	go func() {
		l, err := net.Listen("tcp", nodeCfg.SelfAddr)
		if err != nil {
			log.Fatalf(ctx, "failed to listen: %v", err)
		}
		srv := grpc.NewServer()
		transport.RegisterReplicatorServiceServer(srv, localSrv)
		log.Infof(ctx, "grpc srv %s started on %s", dbName, nodeCfg.SelfAddr)
		err = srv.Serve(l)
		if err != nil {
			log.Fatalf(ctx, "error in local grpc server start on %s", dbName)
		}
	}()
}
