package main

import (
	"fmt"
	"net/http"
	"os"

	pkg "github.com/Gamequic/LivePreviewBackend/pkg/database"
	featuresApi "github.com/Gamequic/LivePreviewBackend/pkg/features"
	"github.com/Gamequic/LivePreviewBackend/utils"
	"github.com/Gamequic/LivePreviewBackend/utils/middlewares"

	"github.com/gorilla/handlers"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

var Logger *zap.Logger

// execute before main
func init() {
	Logger = utils.NewLogger()
	pkg.Logger = Logger
}

func main() {
	// Example usage
	// excelPath := "./data.xlsx"
	// sheetName := "JUN" // Change this to your actual sheet name

	// // Check file exists before proceeding
	// if _, err := os.Stat(excelPath); os.IsNotExist(err) {
	// 	log.Fatalf("❌ File does not exist: %s", excelPath)
	// }

	// // Import sheet
	// rows, err := utils.ImportExcelSheet(excelPath, sheetName)
	// if err != nil {
	// 	log.Fatalf("❌ Could not import Excel sheet: %v", err)
	// }

	// Print all rows
	// fmt.Println("✅ Excel Data:")
	// for i, row := range rows {
	// 	fmt.Printf("Row %d: %v\n", i+1, row)
	// }

	defer Logger.Sync() // Flush log buffer
	utils.Dotconfig()
	pkg.InitDB()
	pkg.InitRedis()

	port := os.Getenv("PORT")
	mainRouter := mux.NewRouter()

	// Middleware for handling errors
	mainRouter.Use(middlewares.ErrorHandler)
	mainRouter.Use(middlewares.GormErrorHandler)

	// Register all API routes
	featuresApi.RegisterSubRoutes(mainRouter)

	// Health check endpoint
	mainRouter.HandleFunc("/checkhealth", utils.CheckHealth).Methods(http.MethodGet)

	// Optional: Handle preflight requests globally
	mainRouter.Methods(http.MethodOptions).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Correct CORS setup
	corsObj := handlers.CORS(
		handlers.AllowedOrigins([]string{"http://localhost:8081"}), // be specific if possible
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Authorization", "Content-Type"}),
	)

	Logger.Info(fmt.Sprintf("🚀 Server running on 0.0.0.0:%s", port))
	err := http.ListenAndServe(fmt.Sprintf(":%s", port), corsObj(mainRouter))
	if err != nil {
		Logger.Fatal("❌ Server failed to start", zap.Error(err))
	}
}
