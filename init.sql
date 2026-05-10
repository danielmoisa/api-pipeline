CREATE TABLE IF NOT EXISTS weather_logs (
    id SERIAL PRIMARY KEY,
    city TEXT,
    raw_response TEXT,
    recorded_at TIMESTAMP DEFAULT NOW()
);
