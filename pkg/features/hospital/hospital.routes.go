package hospital

import (
	"encoding/json"
	"net/http"
	"strconv"

	pieceservice "github.com/Gamequic/LivePreviewBackend/pkg/features/pieces/service"
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
	piecesRouter := router.PathPrefix("/pieces").Subrouter()
	piecesRouter.Use(middlewares.AuthHandler)

	// ValidatorHandler - Update
	usersUpdateValidator := piecesRouter.NewRoute().Subrouter()
	// usersUpdateValidator.Use(middlewares.ValidatorHandler(reflect.TypeOf(userstruct.UpdateUser{})))
	// usersUpdateValidator.Use(middlewares.AuthHandler)
	usersUpdateValidator.HandleFunc("/{id}", update).Methods("PUT")

	// ValidatorHandler - Create
	piecesCreateValidator := piecesRouter.NewRoute().Subrouter()
	// piecesCreateValidator.Use(middlewares.ValidatorHandler(reflect.TypeOf(userstruct.CreateUser{})))
	piecesCreateValidator.HandleFunc("/", create).Methods("POST")

	// Protected functions
	piecesProtected := piecesRouter.NewRoute().Subrouter()
	// piecesProtected.Use(middlewares.ProfilesHandler([]uint{1, 2, 4}))
	piecesProtected.HandleFunc("/", find).Methods("GET")
	piecesProtected.HandleFunc("/{id}", findOne).Methods("GET")
	piecesProtected.HandleFunc("/{id}", delete).Methods("DELETE")
}
