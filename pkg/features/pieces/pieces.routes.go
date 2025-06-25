package pieces

import (
	"encoding/json"
	"net/http"
	"reflect"
	"strconv"

	pieceservice "github.com/Gamequic/LivePreviewBackend/pkg/features/pieces/service"
	userstruct "github.com/Gamequic/LivePreviewBackend/pkg/features/users/struct"
	"github.com/Gamequic/LivePreviewBackend/utils"
	"github.com/Gamequic/LivePreviewBackend/utils/middlewares"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

var logger *zap.Logger

// CRUD

func create(w http.ResponseWriter, r *http.Request) {
	var pieces pieceservice.Pieces

	/*
		This error is alredy been check it on middlewares.ValidatorHandler
		utils/middlewares/validatorHandler.go:29:68
	*/
	json.NewDecoder(r.Body).Decode(&pieces)

	pieceservice.Create(&pieces)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(pieces)
}

func find(w http.ResponseWriter, r *http.Request) {
	//Service
	var pieces []pieceservice.Pieces
	var httpsResponse int = pieceservice.Find(&pieces)

	//Https response
	w.WriteHeader(httpsResponse)
	json.NewEncoder(w).Encode(pieces)
}

func findOne(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		panic(middlewares.GormError{Code: 400, Message: err.Error(), IsGorm: true})
	}
	var pieces pieceservice.Pieces
	var httpsResponse int = pieceservice.FindOne(&pieces, uint(id))
	w.WriteHeader(httpsResponse)
	json.NewEncoder(w).Encode(pieces)
}

func findWithFilters(w http.ResponseWriter, r *http.Request) {
	// Obtener los parámetros de la query string
	filters := map[string]string{
		"identifier": r.URL.Query().Get("identifier"),
		"date":       r.URL.Query().Get("date"),
		"hospital":   r.URL.Query().Get("hospital"),
		"medico":     r.URL.Query().Get("medico"),
		"paciente":   r.URL.Query().Get("paciente"),
	}
	results, status := pieceservice.FindByFilters(filters)
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(results)
}

func update(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idInt, err := strconv.Atoi(vars["id"])
	if err != nil {
		panic(middlewares.GormError{Code: 400, Message: err.Error(), IsGorm: true})
	}
	id := uint(idInt)
	var piece pieceservice.Pieces
	json.NewDecoder(r.Body).Decode(&piece)
	pieceservice.Update(&piece, id)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(piece)
}

func delete(w http.ResponseWriter, r *http.Request) {
	logger = utils.NewLogger()
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		logger.Error("Failed to convert ID to integer", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	pieceservice.Delete(id)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode("Piece deleted successfully")
}

// Register function

func RegisterSubRoutes(router *mux.Router) {
	usersRouter := router.PathPrefix("/pieces").Subrouter()

	// ValidatorHandler - Update
	usersUpdateValidator := usersRouter.NewRoute().Subrouter()
	usersUpdateValidator.Use(middlewares.ValidatorHandler(reflect.TypeOf(userstruct.UpdateUser{})))
	// usersUpdateValidator.Use(middlewares.AuthHandler)
	usersUpdateValidator.HandleFunc("/", update).Methods("PUT")

	// ValidatorHandler - Create
	piecesCreateValidator := usersRouter.NewRoute().Subrouter()
	// piecesCreateValidator.Use(middlewares.ValidatorHandler(reflect.TypeOf(userstruct.CreateUser{})))
	piecesCreateValidator.HandleFunc("/", create).Methods("POST")

	// Protected functions
	usersProtected := usersRouter.NewRoute().Subrouter()
	// usersProtected.Use(middlewares.AuthHandler)
	// usersProtected.Use(middlewares.ProfilesHandler([]uint{1, 2, 4}))
	usersProtected.HandleFunc("/", find).Methods("GET")
	usersProtected.HandleFunc("/search", findWithFilters).Methods("GET")
	usersProtected.HandleFunc("/{id}", findOne).Methods("GET")
	usersProtected.HandleFunc("/{id}", delete).Methods("DELETE")
}
