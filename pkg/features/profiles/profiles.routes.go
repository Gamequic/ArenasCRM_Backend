package profiles

import (
	"encoding/json"
	"net/http"
	profileservice "storegestserver/pkg/features/profiles/service"
	profilestruct "storegestserver/pkg/features/profiles/struct"
	"strconv"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

var logger *zap.Logger

// CRUD

func createProfile(w http.ResponseWriter, r *http.Request) {
	var profile profilestruct.Profile
	if err := json.NewDecoder(r.Body).Decode(&profile); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if err := profileservice.Create(&profile); err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Profile created successfully",
		"id":      profile.ID,
	})
}

func updateProfile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid profile ID", http.StatusBadRequest)
		return
	}

	var profile profilestruct.Profile
	if err := json.NewDecoder(r.Body).Decode(&profile); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if err := profileservice.Update(&profile, id); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Profile updated successfully",
	})
}

func getProfile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid profile ID", http.StatusBadRequest)
		return
	}

	var profile profilestruct.Profile
	if err := profileservice.FindOne(&profile, id); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(profile)
}

func getProfiles(w http.ResponseWriter, r *http.Request) {
	var profiles []profilestruct.Profile
	if err := profileservice.Find(&profiles); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(profiles)
}

func deleteProfile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid profile ID", http.StatusBadRequest)
		return
	}

	if err := profileservice.Delete(id); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Profile deleted successfully",
	})
}

// Register function

func RegisterSubRoutes(router *mux.Router) {
	profilesRouter := router.PathPrefix("/profiles").Subrouter()

	profilesRouter.HandleFunc("/", createProfile).Methods("POST")
	profilesRouter.HandleFunc("/{id}", updateProfile).Methods("PATCH")
	profilesRouter.HandleFunc("/{id}", getProfile).Methods("GET")
	profilesRouter.HandleFunc("/", getProfiles).Methods("GET")
	profilesRouter.HandleFunc("/{id}", deleteProfile).Methods("DELETE")
}
