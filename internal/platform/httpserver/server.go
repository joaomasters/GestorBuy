// Package httpserver monta o *http.Server e o roteador raiz da aplicação.
// Usa o http.ServeMux padrão do Go (com roteamento por método+padrão,
// disponível desde a 1.22) — deliberadamente sem framework de terceiros: no
// tamanho do MVP, o mux da stdlib cobre tudo que os módulos precisam sem
// dependência extra para manter.
package httpserver

import (
	"log/slog"
	"net/http"
	"time"
)

// New monta o servidor HTTP com a cadeia de middlewares globais (log de
// acesso, recover de panic) aplicada sobre o mux fornecido pelo chamador —
// o chamador (main.go) já registrou as rotas de cada módulo nesse mux.
func New(addr string, mux *http.ServeMux, log *slog.Logger) *http.Server {
	handler := recoverMiddleware(log, accessLogMiddleware(log, mux))

	return &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}

func accessLogMiddleware(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r)
		log.Info("http",
			"method", r.Method,
			"path", r.URL.Path,
			"status", sw.status,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})
}

func recoverMiddleware(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Error("panic recuperado", "error", rec, "path", r.URL.Path)
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (sw *statusWriter) WriteHeader(status int) {
	sw.status = status
	sw.ResponseWriter.WriteHeader(status)
}
