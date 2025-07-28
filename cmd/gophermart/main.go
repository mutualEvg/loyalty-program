package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"

	"gofemart/internal/database"
	"gofemart/internal/handlers"
	"gofemart/internal/middleware"
	"gofemart/internal/services"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Configuration flags
	var (
		runAddress        = flag.String("a", ":8080", "server run address")
		databaseURI       = flag.String("d", "", "database connection URI")
		accrualSystemAddr = flag.String("r", "", "accrual system address")
	)
	flag.Parse()

	// Environment variables override flags
	if addr := os.Getenv("RUN_ADDRESS"); addr != "" {
		*runAddress = addr
	}
	if dbURI := os.Getenv("DATABASE_URI"); dbURI != "" {
		*databaseURI = dbURI
	}
	if accrualAddr := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); accrualAddr != "" {
		*accrualSystemAddr = accrualAddr
	}

	// Default database URI if not provided
	if *databaseURI == "" {
		*databaseURI = "postgres://ar11@localhost:5432/gofemart?sslmode=disable"
	}

	// Initialize database
	db, err := database.NewConnection(*databaseURI)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Create tables
	if err := db.CreateTables(); err != nil {
		log.Fatalf("Failed to create tables: %v", err)
	}

	// Initialize services
	userService := services.NewUserService(db)
	orderService := services.NewOrderService(db)
	if *accrualSystemAddr != "" {
		// Use real accrual system if address is provided
		orderService.SetAccrualSystemAddress(*accrualSystemAddr)
	}
	withdrawalService := services.NewWithdrawalService(db)

	// Initialize handlers
	authHandlers := handlers.NewAuthHandlers(userService)
	orderHandlers := handlers.NewOrderHandlers(orderService)
	balanceHandlers := handlers.NewBalanceHandlers(userService, withdrawalService)

	// Setup router
	router := mux.NewRouter()

	// Public routes
	router.HandleFunc("/api/user/register", authHandlers.Register).Methods("POST")
	router.HandleFunc("/api/user/login", authHandlers.Login).Methods("POST")

	// Protected routes
	protected := router.PathPrefix("/api/user").Subrouter()
	protected.Use(middleware.AuthRequired)

	// Order routes
	protected.HandleFunc("/orders", orderHandlers.SubmitOrder).Methods("POST")
	protected.HandleFunc("/orders", orderHandlers.GetUserOrders).Methods("GET")

	// Balance routes
	protected.HandleFunc("/balance", balanceHandlers.GetBalance).Methods("GET")
	protected.HandleFunc("/balance/withdraw", balanceHandlers.WithdrawPoints).Methods("POST")
	protected.HandleFunc("/withdrawals", balanceHandlers.GetWithdrawals).Methods("GET")

	// Add logging middleware
	router.Use(loggingMiddleware)

	log.Printf("Starting Gufermart Loyalty System on %s", *runAddress)
	log.Printf("Database URI: %s", *databaseURI)
	if *accrualSystemAddr != "" {
		log.Printf("Accrual System Address: %s", *accrualSystemAddr)
	}
	log.Println("Available endpoints:")
	log.Println("  POST /api/user/register - User registration")
	log.Println("  POST /api/user/login - User authentication")
	log.Println("  POST /api/user/orders - Submit order number")
	log.Println("  GET  /api/user/orders - Get user orders")
	log.Println("  GET  /api/user/balance - Get user balance")
	log.Println("  POST /api/user/balance/withdraw - Withdraw points")
	log.Println("  GET  /api/user/withdrawals - Get withdrawal history")

	log.Fatal(http.ListenAndServe(*runAddress, router))
}

// loggingMiddleware logs HTTP requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.Method, r.RequestURI, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}
