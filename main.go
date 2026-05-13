package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	// urlExample := "postgres://username:password@localhost:5432/database_name"
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}

	_, err = conn.Exec(context.Background(), `
    CREATE TABLE IF NOT EXISTS weather_logs (
        id SERIAL PRIMARY KEY,
        city TEXT,
        raw_response TEXT,
        recorded_at TIMESTAMP DEFAULT NOW()
    )
`)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to run migrations: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Migrations done")

	defer conn.Close(context.Background())

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// Weather
	http.HandleFunc("/weather", func(w http.ResponseWriter, r *http.Request) {
		// 1. Get city from URL param
		city := r.URL.Query().Get("city")
		if city == "" {
			http.Error(w, "city param is required", http.StatusBadRequest)
			return
		}

		// 2. Build the geocoding URL
		params := url.Values{}
		params.Set("name", city)
		params.Set("count", "1")
		params.Set("language", "en")
		params.Set("format", "json")
		baseURL := "https://geocoding-api.open-meteo.com/v1/search?" + params.Encode()

		// 3. Call the API
		resp, err := http.Get(baseURL)
		if err != nil {
			http.Error(w, "failed to fetch weather", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		// 4. Parse the response body
		var result map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			http.Error(w, fmt.Sprintf("failed to parse response: %v", err), http.StatusInternalServerError)
			return
		}

		// 5. Save to DB
		_, err = conn.Exec(context.Background(),
			"INSERT INTO weather_logs (city, raw_response, recorded_at) VALUES ($1, $2, NOW())",
			city, fmt.Sprintf("%v", result),
		)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to save to db: %v", err), http.StatusInternalServerError)
			return
		}

		// 6. Write response back
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)

	})

	fmt.Println("Server running on :8080")
	http.ListenAndServe(":8080", nil)
}
