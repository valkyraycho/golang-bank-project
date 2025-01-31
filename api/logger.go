package api

import (
	"context"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func GRPCLogger(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	startTime := time.Now()
	res, err := handler(ctx, req)
	duration := time.Since(startTime)

	statusCode := codes.Unknown
	if st, ok := status.FromError(err); ok {
		statusCode = st.Code()
	}

	logger := log.Info()
	if err != nil {
		logger = log.Error().Err(err)
	}

	logger.Str("protocol", "GRPC").
		Str("method", info.FullMethod).
		Int("status_code", int(statusCode)).
		Str("status", statusCode.String()).
		Dur("duration", duration).
		Msg("received a GRPC request")

	return res, err
}

type ResponseRecorder struct {
	http.ResponseWriter
	StatusCode int
	Body       []byte
}

func (recorder *ResponseRecorder) WriteHeader(statusCode int) {
	recorder.StatusCode = statusCode
	recorder.ResponseWriter.WriteHeader(statusCode)
}

func (recorder *ResponseRecorder) Write(body []byte) (int, error) {
	recorder.Body = body
	return recorder.ResponseWriter.Write(body)
}

func HTTPLogger(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		recorder := &ResponseRecorder{
			ResponseWriter: w,
			StatusCode:     http.StatusOK,
		}
		handler.ServeHTTP(recorder, r)
		duration := time.Since(startTime)

		logger := log.Info()
		if recorder.StatusCode >= 400 {
			logger = log.Error().Bytes("body", recorder.Body)
		}

		logger.Str("protocol", "HTTP").
			Str("method", r.Method).
			Str("path", r.RequestURI).
			Int("status_code", recorder.StatusCode).
			Str("status", http.StatusText(recorder.StatusCode)).
			Dur("duration", duration).
			Msg("received a HTTP request")

	})
}
