package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/dariolbs/PSI/internal/config"
	"github.com/dariolbs/PSI/internal/db"
	"github.com/dariolbs/PSI/internal/handlers"
	"github.com/dariolbs/PSI/internal/middleware"
	"github.com/dariolbs/PSI/internal/repository"
)

func main() {
	// Load configuration
	cfg := config.Load()

	log.Printf("Starting server with configuration:")
	log.Printf("  MongoDB URI: %s", cfg.MongoURI)
	log.Printf("  Database Name: %s", cfg.DBName)
	log.Printf("  Server Port: %s", cfg.ServerPort)
	log.Printf("  Frontend URL: %s", cfg.FrontendURL)

	// Connect to MongoDB
	mongodb, err := db.NewMongoDB(cfg.MongoURI, cfg.DBName)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer mongodb.Close()

	log.Println("Successfully connected to MongoDB")

	// Initialize repositories
	userRepo := repository.NewMongoUserRepository(mongodb)
	artistRepo := repository.NewMongoArtistRepository(mongodb)
	albumRepo := repository.NewMongoAlbumRepository(mongodb)
	songRepo := repository.NewMongoSongRepository(mongodb)
	collectionRepo := repository.NewMongoCollectionRepository(mongodb)
	requestRepo := repository.NewMongoVersionRequestRepository(mongodb)
	notificationRepo := repository.NewMongoNotificationRepository(mongodb)
	customListRepo := repository.NewMongoCustomListRepository(mongodb)
	customListItemRepo := repository.NewMongoCustomListItemRepository(mongodb)

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler(mongodb)
	userHandler := handlers.NewUserHandler(userRepo, artistRepo)
	authHandler := handlers.NewAuthHandler(userRepo)
	homeHandler := handlers.NewHomeHandler(albumRepo, artistRepo)
	artistHandler := handlers.NewArtistHandler(artistRepo, albumRepo, songRepo)
	albumHandler := handlers.NewAlbumHandler(albumRepo, artistRepo, songRepo)
	songHandler := handlers.NewSongHandler(songRepo, artistRepo, albumRepo)
	collectionHandler := handlers.NewCollectionHandler(collectionRepo, albumRepo, artistRepo)
	requestHandler := handlers.NewRequestHandler(requestRepo, notificationRepo, albumRepo)
	customListHandler := handlers.NewCustomListHandler(customListRepo, customListItemRepo, albumRepo)

	// Setup HTTP router
	mux := http.NewServeMux()

	// Register routes
	mux.HandleFunc("/health", healthHandler.Health)
	mux.HandleFunc("/health/db", healthHandler.DBHealth)
	mux.HandleFunc("/users", middleware.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			userHandler.ListUsers(w, r)
		case http.MethodPost:
			userHandler.CreateUser(w, r)
		default:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))

	mux.HandleFunc("/users/", middleware.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/users/")
		parts := strings.Split(path, "/")

		if len(parts) == 2 && parts[1] == "favorite-artist" {
			switch r.Method {
			case http.MethodPut:
				userHandler.SetFavoriteArtist(w, r, parts[0])
			case http.MethodDelete:
				userHandler.RemoveFavoriteArtist(w, r, parts[0])
			default:
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
			return
		}

		switch r.Method {
		case http.MethodGet:
			userHandler.GetUser(w, r)
		case http.MethodPut:
			userHandler.UpdateUser(w, r)
		case http.MethodDelete:
			userHandler.DeleteUser(w, r)
		default:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))

	//Auth routes
	mux.HandleFunc("/api/auth/login", authHandler.Login)
	mux.HandleFunc("/api/auth/register", authHandler.Register)

	// Home page routes
	mux.HandleFunc("/api/albums/search", albumHandler.SearchAlbums)
	mux.HandleFunc("/api/albums/featured", homeHandler.FeaturedAlbums)
	mux.HandleFunc("/api/artists/featured", homeHandler.FeaturedArtists)
	mux.HandleFunc("/api/albums/recent", homeHandler.RecentReleases)

	// Search routes
	mux.HandleFunc("/api/artists/search", artistHandler.SearchArtists)

	// Collection routes
	mux.HandleFunc("/api/collections", middleware.RequireAuth(collectionHandler.HandleCollections))
	mux.HandleFunc("/api/collections/", middleware.RequireAuth(collectionHandler.HandleCollectionItem))

	// Detail routes
	mux.HandleFunc("/api/artists/", artistHandler.GetArtist)
	mux.HandleFunc("/api/albums/", albumHandler.GetAlbum)
	mux.HandleFunc("/api/songs/", songHandler.GetSong)
	
	// Version request routes — protected (US12, US13)
	mux.HandleFunc("/api/requests", middleware.RequireAuth(func(w http.ResponseWriter, r *http.Request) {

    	switch r.Method {
    	case http.MethodGet:
        	requestHandler.ListRequests(w, r)
    	case http.MethodPost:
        	requestHandler.CreateRequest(w, r)
    	default:
        	w.WriteHeader(http.StatusMethodNotAllowed)
    	}
	}))

	// Respond to a request — simulates admin action (US14)

	mux.HandleFunc("/api/requests/", middleware.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/respond") && r.Method == http.MethodPost {
			requestHandler.RespondToRequest(w, r)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))

	mux.HandleFunc("/api/lists", middleware.RequireAuth(customListHandler.HandleCustomLists))
	mux.HandleFunc("/api/lists/", middleware.RequireAuth(customListHandler.HandleCustomListItem))

	// Notification routes — protected (US14)
	mux.HandleFunc("/api/notifications", middleware.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
    	switch r.Method {
    	case http.MethodGet:
        	requestHandler.ListNotifications(w, r)
    	default:
        	w.WriteHeader(http.StatusMethodNotAllowed)
    	}
	}))

	mux.HandleFunc("/api/notifications/unread-count", middleware.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
    	switch r.Method {
    	case http.MethodGet:
        	requestHandler.CountUnreadNotifications(w, r)
    	default:
        	w.WriteHeader(http.StatusMethodNotAllowed)
    	}
	}))

	mux.HandleFunc("/api/notifications/mark-all-read", middleware.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
    	switch r.Method {
    	case http.MethodPost:
        	requestHandler.MarkAllNotificationsRead(w, r)
    	default:
        	w.WriteHeader(http.StatusMethodNotAllowed)
    	}
	}))

	mux.HandleFunc("/api/notifications/", middleware.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
    	if strings.HasSuffix(r.URL.Path, "/read") && r.Method == http.MethodPost {
        	requestHandler.MarkNotificationRead(w, r)
        	return
    	}
    	w.WriteHeader(http.StatusNotFound)
	}))

	// CORS middleware
	corsHandler := middleware.CorsMiddleware(mux, cfg.FrontendURL)

	// Create server
	server := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      corsHandler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on port %s", cfg.ServerPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Server is shutting down...")

	// Gracefully shutdown the server
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited properly")
}
