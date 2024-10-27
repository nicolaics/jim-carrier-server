package api

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/nicolaics/jim-carrier/logger"
	"github.com/nicolaics/jim-carrier/service/auth"
	"github.com/nicolaics/jim-carrier/service/currency"
	"github.com/nicolaics/jim-carrier/service/listing"
	"github.com/nicolaics/jim-carrier/service/order"
	"github.com/nicolaics/jim-carrier/service/review"
	"github.com/nicolaics/jim-carrier/service/user"
)

type APIServer struct {
	addr string
	db   *sql.DB
}

func NewAPIServer(addr string, db *sql.DB) *APIServer {
	return &APIServer{
		addr: addr,
		db:   db,
	}
}

func (s *APIServer) Run() error {
	loggerVar := log.New(os.Stdout, "", log.LstdFlags)

	router := mux.NewRouter()
	subrouter := router.PathPrefix("/api/v1").Subrouter()
	subrouterUnprotected := router.PathPrefix("/api/v1").Subrouter()

	userStore := user.NewStore(s.db)
	listingStore := listing.NewStore(s.db)
	orderStore := order.NewStore(s.db)
	reviewStore := review.NewStore(s.db)
	currencyStore := currency.NewStore(s.db)

	userHandler := user.NewHandler(userStore)
	userHandler.RegisterRoutes(subrouter)
	userHandler.RegisterUnprotectedRoutes(subrouterUnprotected)

	listingHandler := listing.NewHandler(listingStore, userStore, currencyStore)
	listingHandler.RegisterRoutes(subrouter)

	orderHandler := order.NewHandler(orderStore, userStore, listingStore, currencyStore)
	orderHandler.RegisterRoutes(subrouter)

	reviewHandler := review.NewHandler(reviewStore, orderStore, listingStore, userStore)
	reviewHandler.RegisterRoutes(subrouter)

	log.Println("Listening on: ", s.addr)

	logMiddleware := logger.NewLogMiddleware(loggerVar)
	router.Use(logMiddleware.Func())

	router.Use(auth.CorsMiddleware())
	subrouter.Use(auth.AuthMiddleware())

	return http.ListenAndServe(s.addr, router)
}
