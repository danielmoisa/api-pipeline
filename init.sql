CREATE TABLE IF NOT EXISTS weather_logs (
  id SERIAL PRIMARY KEY,
  temperature FLOAT,
  windspeed FLOAT,
  weathercode INT,
  recorded_at TIMESTAMP DEFAULT NOW()
);
