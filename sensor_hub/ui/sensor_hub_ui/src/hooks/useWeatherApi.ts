import { useEffect, useState } from "react";
import { useAuth } from "../providers/AuthContext.tsx";

export type DailyForecast = {
  date: string;
  weatherCode: number;
  tempMax: number;
  tempMin: number;
  precipitationProbability: number;
  windSpeedMax: number;
};

export type HourlyForecast = {
  time: string;
  temperature: number;
  apparentTemperature: number;
  precipitationProbability: number;
  weatherCode: number;
  windSpeed: number;
};

export type WeatherForecastData = {
  daily: DailyForecast[];
  hourly: HourlyForecast[];
};

type UseWeatherApiResult = {
  data: WeatherForecastData | null;
  loading: boolean;
  error: string | null;
};

export function useWeatherApi(
  latitude: number,
  longitude: number
): UseWeatherApiResult {
  const { user } = useAuth();
  const [data, setData] = useState<WeatherForecastData | null>(null);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (user === undefined || user === null) return;
    if (!latitude || !longitude) return;

    const controller = new AbortController();

    async function fetchWeather() {
      setLoading(true);
      setError(null);
      try {
        const params = new URLSearchParams({
          latitude: String(latitude),
          longitude: String(longitude),
          daily:
            "weather_code,temperature_2m_max,temperature_2m_min,precipitation_probability_max,wind_speed_10m_max",
          hourly:
            "temperature_2m,apparent_temperature,precipitation_probability,weather_code,wind_speed_10m",
          forecast_days: "7",
          timezone: "auto",
        });

        const url = `https://api.open-meteo.com/v1/forecast?${params.toString()}`;
        const resp = await fetch(url, { signal: controller.signal });
        if (!resp.ok) throw new Error(`HTTP ${resp.status}`);
        const json = await resp.json();

        const dailyRaw = json.daily ?? {};
        const dailyTimes: string[] = dailyRaw.time ?? [];
        const daily: DailyForecast[] = dailyTimes.map(
          (date: string, i: number) => ({
            date,
            weatherCode: dailyRaw.weather_code?.[i] ?? 0,
            tempMax: dailyRaw.temperature_2m_max?.[i] ?? 0,
            tempMin: dailyRaw.temperature_2m_min?.[i] ?? 0,
            precipitationProbability:
              dailyRaw.precipitation_probability_max?.[i] ?? 0,
            windSpeedMax: dailyRaw.wind_speed_10m_max?.[i] ?? 0,
          })
        );

        const hourlyRaw = json.hourly ?? {};
        const hourlyTimes: string[] = hourlyRaw.time ?? [];

        // Filter hourly data to today only — use the first daily date from the
        // API (local tz) rather than UTC to avoid timezone mismatch.
        const todayStr = dailyTimes[0] ?? new Date().toISOString().slice(0, 10);
        const hourly: HourlyForecast[] = hourlyTimes
          .map((time: string, i: number) => ({
            time,
            temperature: hourlyRaw.temperature_2m?.[i] ?? 0,
            apparentTemperature: hourlyRaw.apparent_temperature?.[i] ?? 0,
            precipitationProbability:
              hourlyRaw.precipitation_probability?.[i] ?? 0,
            weatherCode: hourlyRaw.weather_code?.[i] ?? 0,
            windSpeed: hourlyRaw.wind_speed_10m?.[i] ?? 0,
          }))
          .filter((h: HourlyForecast) => h.time.startsWith(todayStr));

        setData({ daily, hourly });
      } catch (err: unknown) {
        const name = err instanceof Error ? err.name : "";
        if (name === "AbortError") return;
        const message = err instanceof Error ? err.message : String(err);
        setError(message);
        setData(null);
      } finally {
        setLoading(false);
      }
    }

    fetchWeather();
    return () => controller.abort();
  }, [latitude, longitude, user]);

  return { data, loading, error };
}

export default useWeatherApi;
