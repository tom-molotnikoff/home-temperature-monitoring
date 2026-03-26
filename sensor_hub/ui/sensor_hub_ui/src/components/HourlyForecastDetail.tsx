import Box from "@mui/material/Box";
import Paper from "@mui/material/Paper";
import Typography from "@mui/material/Typography";
import WaterDropOutlined from "@mui/icons-material/WaterDropOutlined";
import AirOutlined from "@mui/icons-material/AirOutlined";
import { getWeatherInfo } from "../tools/weatherIcons.ts";
import type { HourlyForecast } from "../hooks/useWeatherApi.ts";

type HourlyForecastDetailProps = {
  hours: HourlyForecast[];
  compact?: boolean;
};

function formatHour(timeStr: string): string {
  const date = new Date(timeStr);
  return date.toLocaleTimeString(undefined, {
    hour: "2-digit",
    minute: "2-digit",
    hour12: false,
  });
}

export default function HourlyForecastDetail({
  hours,
  compact,
}: HourlyForecastDetailProps) {
  if (hours.length === 0) {
    return (
      <Typography variant="body2" color="text.secondary" sx={{ p: 2 }}>
        No hourly data available for today.
      </Typography>
    );
  }

  return (
    <Box
      sx={{
        display: "flex",
        gap: 1,
        overflowX: "auto",
        pb: 1,
        "&::-webkit-scrollbar": { height: 6 },
        "&::-webkit-scrollbar-thumb": {
          borderRadius: 3,
          bgcolor: "action.disabled",
        },
      }}
    >
      {hours.map((h) => {
        const { icon: WeatherIcon } = getWeatherInfo(h.weatherCode);
        return (
          <Paper
            key={h.time}
            variant="outlined"
            sx={{
              p: compact ? 0.75 : 1,
              minWidth: compact ? 48 : 56,
              textAlign: "center",
              flex: compact ? "0 0 auto" : "1 1 0",
              borderRadius: compact ? 3 : 2,
            }}
          >
            <Typography variant="caption" fontWeight="bold">
              {formatHour(h.time)}
            </Typography>

            <Box sx={{ my: compact ? 0.25 : 0.5 }}>
              <WeatherIcon sx={{ fontSize: compact ? 16 : 20, color: "primary.main" }} />
            </Box>

            <Typography variant={compact ? "caption" : "body2"} fontWeight="bold">
              {Math.round(h.temperature)}°
            </Typography>
            {!compact && (
              <Typography variant="caption" color="text.secondary" display="block">
                Feels {Math.round(h.apparentTemperature)}°
              </Typography>
            )}

            <Box
              sx={{
                display: "flex",
                alignItems: "center",
                justifyContent: "center",
                gap: 0.3,
                mt: 0.5,
              }}
            >
              <WaterDropOutlined sx={{ fontSize: compact ? 10 : 12, color: "info.main" }} />
              <Typography variant="caption" fontSize={compact ? "0.55rem" : "0.65rem"}>
                {h.precipitationProbability}%
              </Typography>
            </Box>

            {!compact && (
              <Box
                sx={{
                  display: "flex",
                  alignItems: "center",
                  justifyContent: "center",
                  gap: 0.3,
                }}
              >
                <AirOutlined sx={{ fontSize: 12, color: "text.secondary" }} />
                <Typography variant="caption" fontSize="0.65rem">
                  {Math.round(h.windSpeed)}
                </Typography>
              </Box>
            )}
          </Paper>
        );
      })}
    </Box>
  );
}
