package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type API struct {
	playerHandler *PlayerHandler
}

type PlayerHandler struct {
	Conn *sql.DB
}

type PlayerResponse struct {
	ID          string    `json:"id"`
	MatchID     uuid.UUID `json:"matchId"`
	MentionName string    `json:"mentionName"`
	NumWins     int       `json:"numWins"`
	NumLosses   int       `json:"numLosses"`
	MMR         float64   `json:"mmr"`
	IsInGame    bool      `json:"isInGame"`
	IsAdmin     bool      `json:"isAdmin"`
	DiscordID   int       `json:"discordId"`
	Name        string    `json:"name"`
}

func NewAPI(connStr string) *API {
	connection, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Test the connection
	if err := connection.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	playerHandler := &PlayerHandler{Conn: connection}
	return &API{playerHandler: playerHandler}
}

func (api *API) StartAPI(port string) {
	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/players", api.corsMiddleware(api.getPlayers))
	mux.HandleFunc("/api/players/", api.corsMiddleware(api.getPlayerByDiscordID))
	mux.HandleFunc("/api/leaderboard", api.corsMiddleware(api.getLeaderboard))

	// Health check
	mux.HandleFunc("/health", api.corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))

	fmt.Printf("API server starting on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

func (api *API) corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

func (api *API) getPlayers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	players, err := api.playerHandler.GetAllPlayers()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching players: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(players)
}

func (api *API) getPlayerByDiscordID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract Discord ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/players/")
	discordID, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, "Invalid Discord ID", http.StatusBadRequest)
		return
	}

	player, err := api.playerHandler.GetByDiscordID(discordID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Player not found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("Error fetching player: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(player)
}

func (api *API) getLeaderboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	players, err := api.playerHandler.GetLeaderboard()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching leaderboard: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(players)
}

// Database methods
func (handler *PlayerHandler) GetAllPlayers() ([]PlayerResponse, error) {
	rows, err := handler.Conn.Query(`SELECT "id", "Name", "MMR", "Wins", "Losses", "MatchUID", "DiscordId" FROM rocketleague ORDER BY "MMR" DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var players []PlayerResponse
	for rows.Next() {
		var index int
		var name string
		var mmr float64
		var numWins int
		var numLosses int
		var matchID uuid.UUID
		var discordID int

		if err := rows.Scan(&index, &name, &mmr, &numWins, &numLosses, &matchID, &discordID); err != nil {
			return nil, err
		}

		player := PlayerResponse{
			ID:          name,
			Name:        name,
			MatchID:     matchID,
			MentionName: name,
			NumWins:     numWins,
			NumLosses:   numLosses,
			MMR:         mmr,
			IsInGame:    matchID != uuid.Nil,
			IsAdmin:     false,
			DiscordID:   discordID,
		}
		players = append(players, player)
	}

	return players, rows.Err()
}

func (handler *PlayerHandler) GetByDiscordID(discordID int) (*PlayerResponse, error) {
	row := handler.Conn.QueryRow(`SELECT "id", "Name", "MMR", "Wins", "Losses", "MatchUID", "DiscordId" FROM rocketleague WHERE "DiscordId" = $1`, discordID)

	var index int
	var name string
	var mmr float64
	var numWins int
	var numLosses int
	var matchID uuid.UUID
	var dbDiscordID int

	if err := row.Scan(&index, &name, &mmr, &numWins, &numLosses, &matchID, &dbDiscordID); err != nil {
		return nil, err
	}

	player := &PlayerResponse{
		ID:          name,
		Name:        name,
		MatchID:     matchID,
		MentionName: name,
		NumWins:     numWins,
		NumLosses:   numLosses,
		MMR:         mmr,
		IsInGame:    matchID != uuid.Nil,
		IsAdmin:     false,
		DiscordID:   dbDiscordID,
	}

	return player, nil
}

func (handler *PlayerHandler) GetLeaderboard() ([]PlayerResponse, error) {
	rows, err := handler.Conn.Query(`SELECT "id", "Name", "MMR", "Wins", "Losses", "MatchUID", "DiscordId" FROM rocketleague ORDER BY "MMR" DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var players []PlayerResponse
	for rows.Next() {
		var index int
		var name string
		var mmr float64
		var numWins int
		var numLosses int
		var matchID uuid.UUID
		var discordID int

		if err := rows.Scan(&index, &name, &mmr, &numWins, &numLosses, &matchID, &discordID); err != nil {
			return nil, err
		}

		player := PlayerResponse{
			ID:          name,
			Name:        name,
			MatchID:     matchID,
			MentionName: name,
			NumWins:     numWins,
			NumLosses:   numLosses,
			MMR:         mmr,
			IsInGame:    matchID != uuid.Nil,
			IsAdmin:     false,
			DiscordID:   discordID,
		}
		players = append(players, player)
	}

	return players, rows.Err()
}
