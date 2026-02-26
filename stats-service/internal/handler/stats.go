package handler

import (
    "encoding/json"
    "net/http"
    "stats-service/internal/repository"
    "strconv"
    "github.com/gorilla/mux"
)

type StatsHandler struct {
    repo *repository.PostgresRepo
}

func NewStatsHandler(repo *repository.PostgresRepo) *StatsHandler {
    return &StatsHandler{repo: repo}
}

func (h *StatsHandler) GetTodayVisits(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    clubID, _ := strconv.Atoi(vars["clubId"])
    
    count, err := h.repo.GetTodayVisits(clubID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    json.NewEncoder(w).Encode(map[string]int{"todayVisits": count})
}

func (h *StatsHandler) GetTopMembers(w http.ResponseWriter, r *http.Request) {
    members, err := h.repo.GetTopActiveMembers(10)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    json.NewEncoder(w).Encode(members)
}

func (h *StatsHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
    json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}