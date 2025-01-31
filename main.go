package main

import (
	"context"
	"log"
	"net"
	"net/http"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/valkyraycho/bank_project/api"
	db "github.com/valkyraycho/bank_project/db/sqlc"
	"github.com/valkyraycho/bank_project/pb"
	"github.com/valkyraycho/bank_project/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
)

const migrationURL = "file://db/migration"

func main() {
	cfg, err := utils.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load environment variables")
	}

	runDBMigrations(migrationURL, cfg.DBSource)

	connPool, err := pgxpool.New(context.Background(), cfg.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db: ", err)
	}

	store := db.NewStore(connPool)
	go runHTTPServer(context.Background(), cfg, store)
	runGRPCServer(context.Background(), cfg, store)
}

func runDBMigrations(migrationURL, dbsource string) {
	migration, err := migrate.New(migrationURL, dbsource)
	if err != nil {
		log.Fatal("cannot create new migration instance: ", err)
	}
	if err := migration.Up(); err != nil {
		if err == migrate.ErrNoChange {
			log.Println("database already migrated to the latest version")
			return
		}
		log.Fatal("failed to migrate")
		return
	}
	log.Println("database migrated successfully")
}

func runGRPCServer(ctx context.Context, cfg utils.Config, store db.Store) {
	server, err := api.NewServer(cfg, store)
	if err != nil {
		log.Fatal("failed to create server")
	}
	grpcServer := grpc.NewServer()

	pb.RegisterBankServiceServer(grpcServer, server)
	reflection.Register(grpcServer)

	lis, err := net.Listen("tcp", cfg.GRPCServerAddress)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("start gRPC server at %s", cfg.GRPCServerAddress)
	grpcServer.Serve(lis)
}

func runHTTPServer(ctx context.Context, cfg utils.Config, store db.Store) {
	server, err := api.NewServer(cfg, store)
	if err != nil {
		log.Fatal("failed to create server")
	}

	mux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				EmitUnpopulated: true,
				UseProtoNames:   true,
			},
			UnmarshalOptions: protojson.UnmarshalOptions{
				DiscardUnknown: true,
			},
		}),
	)

	if err := pb.RegisterBankServiceHandlerServer(ctx, mux, server); err != nil {
		log.Fatal("failed to register http handler server")
	}

	log.Printf("start http server at %s", cfg.HTTPServerAddress)
	http.ListenAndServe(cfg.HTTPServerAddress, mux)
}
