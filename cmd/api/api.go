package api

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gorilla/mux"
	"github.com/nicolaics/jim-carrier-server/logger"
	"github.com/nicolaics/jim-carrier-server/service/auth"
	"github.com/nicolaics/jim-carrier-server/service/auth/jwt"
	"github.com/nicolaics/jim-carrier-server/service/bank"
	"github.com/nicolaics/jim-carrier-server/service/currency"
	"github.com/nicolaics/jim-carrier-server/service/fcm"
	"github.com/nicolaics/jim-carrier-server/service/listing"
	"github.com/nicolaics/jim-carrier-server/service/order"
	"github.com/nicolaics/jim-carrier-server/service/review"
	"github.com/nicolaics/jim-carrier-server/service/user"
)

type APIServer struct {
	addr string
	db   *sql.DB
	router *mux.Router
	bucket *s3.S3
}

func NewAPIServer(addr string, db *sql.DB, router *mux.Router, bucket *s3.S3) *APIServer {
	return &APIServer{
		addr: addr,
		db:   db,
		router: router,
		bucket: bucket,
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

	userHandler := user.NewHandler(userStore, s.bucket)
	userHandler.RegisterRoutes(subrouter)
	userHandler.RegisterUnprotectedRoutes(subrouterUnprotected)

	listingHandler := listing.NewHandler(listingStore, userStore, currencyStore, reviewStore,
										bankDetailStore, orderStore, fcmStore, s.bucket)
	listingHandler.RegisterRoutes(subrouter)

	orderHandler := order.NewHandler(orderStore, userStore, listingStore, currencyStore, fcmStore, 
									bankDetailStore, s.bucket)
	orderHandler.RegisterRoutes(subrouter)

	reviewHandler := review.NewHandler(reviewStore, orderStore, listingStore, userStore)
	reviewHandler.RegisterRoutes(subrouter)

	bankDetailHandler := bank.NewHandler(bankDetailStore, userStore)
	bankDetailHandler.RegisterRoutes(subrouter)

	log.Println("Listening on: ", s.addr)

	logMiddleware := logger.NewLogMiddleware(loggerVar)
	s.router.Use(logMiddleware.Func())

	s.router.Use(auth.CorsMiddleware())
	subrouter.Use(jwt.JWTMiddleware())

	return http.ListenAndServe(s.addr, s.router)
}
