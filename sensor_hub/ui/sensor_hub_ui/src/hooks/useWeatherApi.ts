import { useCallback, useEffect, useMemo, useState } from "react";

type WeatherPoint = {
  time: string;
  [key: string]: string | number | undefined;
};

type UseWeatherApiResult = {
  data: WeatherPoint[];
  loading: boolean;
  error: string | null;
  refetch: () => void;
};

type HookOptions = {
  hourly?: string[];
  days?: number;
  startDate?: string | Date | null;
  endDate?: string | Date | null;
  timezone?: string;
};

function formatDate(d: Date) {
  const yyyy = d.getUTCFullYear();
  const mm = String(d.getUTCMonth() + 1).padStart(2, "0");
  const dd = String(d.getUTCDate()).padStart(2, "0");
  return `${yyyy}-${mm}-${dd}`;
}

export function useWeatherApi(
  latitude: number,
  longitude: number,
  opts?: HookOptions
): UseWeatherApiResult {
  const hourlyVars = useMemo(
    () =>
      opts?.hourly ?? ["apparent_temperature", "temperature_2m", "uv_index"],
    [opts?.hourly?.join(",")]
  );
  const days = opts?.days ?? 1;
  const timezone = opts?.timezone ?? "auto";

  const startDep = opts?.startDate
    ? opts.startDate instanceof Date
      ? opts.startDate.toISOString()
      : String(opts.startDate)
    : "";
  const endDep = opts?.endDate
    ? opts.endDate instanceof Date
      ? opts.endDate.toISOString()
      : String(opts.endDate)
    : "";

  const [startDateStr, endDateStr] = useMemo(() => {
    const optStart = opts?.startDate ?? null;
    const optEnd = opts?.endDate ?? null;

    const end = optEnd ? new Date(optEnd) : new Date();
    const requestEnd = new Date(end);
    requestEnd.setUTCDate(requestEnd.getUTCDate() - 1);

    if (optStart) {
      const start = new Date(optStart);
      if (start > requestEnd) start.setTime(requestEnd.getTime());
      return [formatDate(start), formatDate(requestEnd)];
    }

    const start = new Date(end);
    start.setUTCDate(start.getUTCDate() - Math.max(1, days));
    if (start > requestEnd) start.setTime(requestEnd.getTime());
    return [formatDate(start), formatDate(requestEnd)];
  }, [days, startDep, endDep]);

  const [data, setData] = useState<WeatherPoint[]>([]);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);
  const [counter, setCounter] = useState<number>(0); // used to trigger refetch

  const refetch = useCallback(() => setCounter((c) => c + 1), []);

  useEffect(() => {
    const controller = new AbortController();
    async function fetchWeather() {
      setLoading(true);
      setError(null);
      try {
        const params = new URLSearchParams({
          latitude: String(latitude),
          longitude: String(longitude),
          hourly: hourlyVars.join(","),
          timezone,
          start_date: startDateStr,
          end_date: endDateStr,
        });

        const url = `https://api.open-meteo.com/v1/forecast?${params.toString()}`;
        const resp = await fetch(url, { signal: controller.signal });
        if (!resp.ok) throw new Error(`HTTP ${resp.status}`);
        const json = await resp.json();

        const hourly = json.hourly ?? {};
        const times: string[] = hourly.time ?? [];

        const out: WeatherPoint[] = times.map((t: string, i: number) => {
          const point: WeatherPoint = { time: t };
          for (const v of hourlyVars) {
            const arr = hourly[v];
            if (Array.isArray(arr)) {
              point[v] = arr[i];
            } else {
              point[v] = undefined;
            }
          }
          return point;
        });

        setData(out);
      } catch (err: unknown) {
        const name = err instanceof Error ? err.name : "";
        if (name === "AbortError") return;
        const message = err instanceof Error ? err.message : String(err);
        setError(message);
        setData([]);
      } finally {
        setLoading(false);
      }
    }

    fetchWeather();

    return () => controller.abort();
  }, [
    latitude,
    longitude,
    timezone,
    counter,
    startDateStr,
    endDateStr,
    hourlyVars.join(","),
  ]);

  return { data, loading, error, refetch };
}

export default useWeatherApi;
