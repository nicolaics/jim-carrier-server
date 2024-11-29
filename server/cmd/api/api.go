package api

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/nicolaics/jim-carrier/logger"
	"github.com/nicolaics/jim-carrier/service/auth"
	"github.com/nicolaics/jim-carrier/service/auth/jwt"
	"github.com/nicolaics/jim-carrier/service/bank"
	"github.com/nicolaics/jim-carrier/service/currency"
	"github.com/nicolaics/jim-carrier/service/fcm"
	"github.com/nicolaics/jim-carrier/service/listing"
	"github.com/nicolaics/jim-carrier/service/order"
	"github.com/nicolaics/jim-carrier/service/review"
	"github.com/nicolaics/jim-carrier/service/user"
)

type APIServer struct {
	addr string
	db   *sql.DB
	router *mux.Router
}

func NewAPIServer(addr string, db *sql.DB, router *mux.Router) *APIServer {
	return &APIServer{
		addr: addr,
		db:   db,
		router: router,
	}
}

func (s *APIServer) Run() error {
	loggerVar := log.New(os.Stdout, "", log.LstdFlags)

	subrouter := s.router.PathPrefix("/api/v1").Subrouter()
	subrouterUnprotected := s.router.PathPrefix("/api/v1").Subrouter()

	userStore := user.NewStore(s.db)
	listingStore := listing.NewStore(s.db)
	orderStore := order.NewStore(s.db)
	reviewStore := review.NewStore(s.db)
	currencyStore := currency.NewStore(s.db)
	fcmStore := fcm.NewStore(s.db)
	bankDetailStore := bank.NewStore(s.db)

	userHandler := user.NewHandler(userStore)
	userHandler.RegisterRoutes(subrouter)
	userHandler.RegisterUnprotectedRoutes(subrouterUnprotected)

	listingHandler := listing.NewHandler(listingStore, userStore, currencyStore, reviewStore, bankDetailStore, orderStore, fcmStore)
	listingHandler.RegisterRoutes(subrouter)

	orderHandler := order.NewHandler(orderStore, userStore, listingStore, currencyStore, fcmStore, bankDetailStore)
	orderHandler.RegisterRoutes(subrouter)

	reviewHandler := review.NewHandler(reviewStore, orderStore, listingStore, userStore)
	reviewHandler.RegisterRoutes(subrouter)

	log.Println("Listening on: ", s.addr)

	logMiddleware := logger.NewLogMiddleware(loggerVar)
	s.router.Use(logMiddleware.Func())

	s.router.Use(auth.CorsMiddleware())
	subrouter.Use(jwt.JWTMiddleware())

	return http.ListenAndServe(s.addr, s.router)
}
