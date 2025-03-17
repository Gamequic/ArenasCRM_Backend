package system

import (
	"fmt"
	"net/http"
	"time"

	systemservice "github.com/Gamequic/LivePreviewBackend/pkg/features/system/service"
	"github.com/Gamequic/LivePreviewBackend/utils"
	"github.com/Gamequic/LivePreviewBackend/utils/middlewares"
	"github.com/gorilla/mux"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/net"
)

var Logger = utils.NewLogger()

func SystemMetricsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := utils.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		Logger.Error(fmt.Sprintf("No se pudo actualizar a WebSocket: %v", err))
		return
	}
	defer conn.Close()

	firstMessage := true
	clientActive := make(chan bool)
	go func() {
		for {
			_, message, err := conn.ReadMessage()
			if firstMessage {
				firstMessage = false
				clientActive <- true
				userID := middlewares.ValidateUser(string(message))
				if userID == -1 {
					Logger.Info("User not valid for system metrics")
					conn.Close()
					return
				} else {
					Logger.Info(fmt.Sprintf("User %v connected in system metrics", userID))
				}
			}
			if err != nil {
				Logger.Info(fmt.Sprintf("Client desconection from system metrics: %v", err))
				clientActive <- false
				return
			}
			if string(message) == `{"type":"ping"}` {
				clientActive <- true
			}
		}
	}()

	var prevNetStats []net.IOCountersStat
	var prevDiskStats map[string]disk.IOCountersStat

	timeout := time.NewTimer(40 * time.Second)
	for {
		select {
		case <-clientActive:
			timeout.Reset(40 * time.Second)

		case <-timeout.C:
			Logger.Info("Client is not active, closing conection from system metrics.")
			conn.Close()
			return

		default:
			stats, err := systemservice.GetSystemMetrics(&prevNetStats, &prevDiskStats)
			if err != nil {
				Logger.Error(fmt.Sprintf("Error obteniendo métricas: %v", err))
				continue
			}

			err = conn.WriteJSON(stats)
			if err != nil {
				Logger.Info(fmt.Sprintf("Error enviando datos: %v", err))
				return
			}

			time.Sleep(1 * time.Second)
		}
	}
}

// Registrar las rutas
func RegisterSubRoutes(router *mux.Router) {
	systemRouter := router.PathPrefix("/system").Subrouter()
	systemRouter.HandleFunc("/metrics", SystemMetricsHandler).Methods("GET")
}
