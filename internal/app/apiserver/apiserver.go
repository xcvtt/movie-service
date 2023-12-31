package apiserver

import (
	"fmt"
	"github.com/exclide/movie-service/internal/app/directors"
	"github.com/exclide/movie-service/internal/app/movies"
	"github.com/exclide/movie-service/internal/app/store"
	"github.com/exclide/movie-service/internal/app/users"
	"github.com/exclide/movie-service/pkg/httpformat"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/golang-jwt/jwt"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

type ApiServer struct {
	config Config
	logger *logrus.Logger
	router chi.Router
	store  *store.Store
}

func NewServer(config Config) *ApiServer {
	return &ApiServer{
		config: config,
		logger: logrus.New(),
		router: chi.NewRouter(),
	}
}

func (s *ApiServer) Start() error {
	if err := s.configureLogger(); err != nil {
		return err
	}

	if err := s.configureStore(); err != nil {
		return err
	}

	s.configureRouter()

	s.logger.Info("starting api server")

	return http.ListenAndServe(s.config.Port, s.router)
}

func (s *ApiServer) configureLogger() error {
	level, err := logrus.ParseLevel(s.config.LogLevel)
	if err != nil {
		return err
	}

	s.logger.SetLevel(level)

	return nil
}

func (s *ApiServer) Root(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte("hello world!"))
	if err != nil {
		return
	}
}

func contentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func authorize(next http.Handler) http.Handler {
	const BearerSchema = "Bearer "

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		if len(tokenString) <= len(BearerSchema) {
			httpformat.Respond(w, r, http.StatusUnauthorized, "invalid token provided")
			return
		}

		//splitToken := strings.Split(tokenString, "Bearer ")

		token, err := jwt.Parse(tokenString[7:], func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}

			key := []byte("SecretKey")

			return key, nil
		})

		if err != nil {
			httpformat.Error(w, r, http.StatusUnauthorized, err)
			return
		}

		if !token.Valid {
			httpformat.Respond(w, r, http.StatusUnauthorized, "invalid token")
			return
		}

		claims := token.Claims.(jwt.MapClaims)

		fmt.Println(claims["sub"])
		fmt.Println(claims["exp"])

		next.ServeHTTP(w, r)
	})
}

func (s *ApiServer) configureRouter() {
	// A good base middleware stack
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RealIP)
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	s.router.Use(middleware.Timeout(60 * time.Second))
	s.router.Use(contentType)
	s.router.Get("/", s.Root)

	movieRepo := movies.NewRepository(s.store)
	movieServ := movies.NewService(movieRepo)
	movieHandler := movies.NewHandler(movieServ)

	dirRepo := directors.NewRepository(s.store)
	dirServ := directors.NewService(dirRepo)
	dirHandler := directors.NewDirectorHandler(dirServ)

	userRepo := users.NewRepository(s.store)
	userServ := users.NewService(userRepo)
	userHandler := users.NewUserHandler(userServ)

	directors.Route(s.router, dirHandler)
	movies.Route(s.router, movieHandler)
	users.Route(s.router, userHandler, authorize)

	s.router.Post("/api/v1/login", userHandler.Login)
}

func (s *ApiServer) configureStore() error {
	st := store.New(s.config.Store)

	if err := st.Open(); err != nil {
		return err
	}

	s.store = st

	return nil
}
