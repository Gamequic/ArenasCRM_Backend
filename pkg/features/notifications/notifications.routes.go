package notifications

import (
	"encoding/json"
	"net/http"
	"reflect"
	"strconv"

	notificationservice "github.com/Gamequic/LivePreviewBackend/pkg/features/notifications/service"
	notificatonstruct "github.com/Gamequic/LivePreviewBackend/pkg/features/notifications/struct"
	"github.com/Gamequic/LivePreviewBackend/utils/middlewares"

	"github.com/gorilla/mux"
)

// CRUD

func create(w http.ResponseWriter, r *http.Request) {
	var notification notificationservice.Notifications
	/*
		This error is alredy been check it on middlewares.ValidatorHandler
		utils/middlewares/validatorHandler.go:29:68
	*/
	json.NewDecoder(r.Body).Decode(&notification)
	userId := r.Context().Value(middlewares.UserIDKey).(int)

	notificationservice.Create(&notification, userId)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(notification)
}

func find(w http.ResponseWriter, r *http.Request) {
	//Service
	var notifications []notificationservice.Notifications
	var httpsResponse int = notificationservice.Find(&notifications)

	//Https response
	w.WriteHeader(httpsResponse)
	json.NewEncoder(w).Encode(notifications)
}

func MarkAsSeen(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		panic(middlewares.GormError{Code: 400, Message: err.Error(), IsGorm: true})
	}
	var httpsResponse int = notificationservice.MarkAsSeen(id)
	w.WriteHeader(httpsResponse)
	json.NewEncoder(w).Encode("Notification marked as seen")
}

// Register function

func RegisterSubRoutes(router *mux.Router) {
	notificationsRouter := router.PathPrefix("/notifications").Subrouter()

	// Protected functions
	notificationsProtected := notificationsRouter.NewRoute().Subrouter()
	notificationsProtected.Use(middlewares.AuthHandler)
	notificationsProtected.Use(middlewares.ProfilesHandler([]uint{1, 2, 7}))
	notificationsProtected.HandleFunc("/", find).Methods("GET")
	notificationsProtected.HandleFunc("/markAsSeen/{id}", MarkAsSeen).Methods("PUT")

	// ValidatorHandler - Create
	notificationsCreateValidator := notificationsProtected.NewRoute().Subrouter()
	notificationsCreateValidator.Use(middlewares.ValidatorHandler(reflect.TypeOf(notificatonstruct.NotificationCreate{})))
	notificationsCreateValidator.HandleFunc("/", create).Methods("POST")

	// Websocket
	notificationsRouter.HandleFunc("/live", notificationservice.NotificationWebSocketEndpoint).Methods("GET")
}
