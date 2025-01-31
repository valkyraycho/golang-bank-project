package main

import (
	"context"
	"net"
	"net/http"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

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
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	cfg, err := utils.LoadConfig(".")
	if err != nil {
		log.Fatal().Msgf("cannot load environment variables: %s", err)
	}

	runDBMigrations(migrationURL, cfg.DBSource)

	connPool, err := pgxpool.New(context.Background(), cfg.DBSource)
	if err != nil {
		log.Fatal().Msgf("cannot connect to db: %s", err)
	}

	store := db.NewStore(connPool)
	go runHTTPServer(context.Background(), cfg, store)
	runGRPCServer(context.Background(), cfg, store)
}

func runDBMigrations(migrationURL, dbsource string) {
	migration, err := migrate.New(migrationURL, dbsource)
	if err != nil {
		log.Fatal().Msgf("cannot create new migration instance: %s", err)
	}
	if err := migration.Up(); err != nil {
		if err == migrate.ErrNoChange {
			log.Info().Msg("database already migrated to the latest version")
			return
		}
		log.Fatal().Msgf("failed to migrate: %s", err)
		return
	}
	log.Info().Msg("database migrated successfully")
}

func runGRPCServer(ctx context.Context, cfg utils.Config, store db.Store) {
	server, err := api.NewServer(cfg, store)
	if err != nil {
		log.Fatal().Msgf("failed to create server: %s", err)
	}

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(api.GRPCLogger))

	pb.RegisterBankServiceServer(grpcServer, server)
	reflection.Register(grpcServer)

	lis, err := net.Listen("tcp", cfg.GRPCServerAddress)
	if err != nil {
		log.Fatal().Msgf("failed to listen: %v", err)
	}
	log.Info().Msgf("start gRPC server at %s", cfg.GRPCServerAddress)
	grpcServer.Serve(lis)
}

func runHTTPServer(ctx context.Context, cfg utils.Config, store db.Store) {
	server, err := api.NewServer(cfg, store)
	if err != nil {
		log.Fatal().Msgf("failed to create server: %s", err)
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
		log.Fatal().Msg("failed to register http handler server")
	}

	log.Info().Msgf("start http server at %s", cfg.HTTPServerAddress)
	http.ListenAndServe(cfg.HTTPServerAddress, api.HTTPLogger(mux))
}
