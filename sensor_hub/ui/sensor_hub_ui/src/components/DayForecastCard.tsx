import Paper from "@mui/material/Paper";
import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import WaterDropOutlined from "@mui/icons-material/WaterDropOutlined";
import AirOutlined from "@mui/icons-material/AirOutlined";
import { getWeatherInfo } from "../tools/weatherIcons.ts";
import type { DailyForecast } from "../hooks/useWeatherApi.ts";

type DayForecastCardProps = {
  day: DailyForecast;
  isToday: boolean;
  compact?: boolean;
};

function formatDayName(dateStr: string): string {
  const date = new Date(dateStr + "T00:00:00");
  return date.toLocaleDateString(undefined, { weekday: "short" });
}

function formatShortDate(dateStr: string): string {
  const date = new Date(dateStr + "T00:00:00");
  return date.toLocaleDateString(undefined, { day: "numeric", month: "short" });
}

export default function DayForecastCard({ day, isToday, compact }: DayForecastCardProps) {
  const { icon: WeatherIcon, label } = getWeatherInfo(day.weatherCode);

  if (compact) {
    return (
      <Paper
        elevation={isToday ? 3 : 1}
        sx={{
          px: 1.5,
          py: 1,
          minWidth: 72,
          textAlign: "center",
          border: isToday ? 2 : 0,
          borderColor: "primary.main",
          borderRadius: 3,
          flex: "0 0 auto",
        }}
      >
        <Typography variant="caption" fontWeight="bold" display="block">
          {isToday ? "Today" : formatDayName(day.date)}
        </Typography>
        <WeatherIcon sx={{ fontSize: 22, color: "primary.main", my: 0.25 }} />
        <Typography variant="caption" fontWeight="bold" display="block">
          {Math.round(day.tempMax)}°/{Math.round(day.tempMin)}°
        </Typography>
        <Box sx={{ display: "flex", alignItems: "center", justifyContent: "center", gap: 0.3 }}>
          <WaterDropOutlined sx={{ fontSize: 10, color: "info.main" }} />
          <Typography variant="caption" fontSize="0.6rem">
            {day.precipitationProbability}%
          </Typography>
        </Box>
      </Paper>
    );
  }

  return (
    <Paper
      elevation={isToday ? 3 : 1}
      sx={{
        p: 1.5,
        minWidth: 100,
        textAlign: "center",
        border: isToday ? 2 : 0,
        borderColor: "primary.main",
        borderRadius: 2,
        flex: "1 1 0",
      }}
    >
      <Typography variant="subtitle2" fontWeight="bold">
        {isToday ? "Today" : formatDayName(day.date)}
      </Typography>
      <Typography variant="caption" color="text.secondary">
        {formatShortDate(day.date)}
      </Typography>

      <Box sx={{ my: 1 }}>
        <WeatherIcon sx={{ fontSize: 36, color: "primary.main" }} />
      </Box>
      <Typography variant="caption" display="block" color="text.secondary">
        {label}
      </Typography>

      <Typography variant="body2" fontWeight="bold" sx={{ mt: 1 }}>
        {Math.round(day.tempMax)}° / {Math.round(day.tempMin)}°
      </Typography>

      <Box
        sx={{
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          gap: 0.5,
          mt: 0.5,
        }}
      >
        <WaterDropOutlined sx={{ fontSize: 14, color: "info.main" }} />
        <Typography variant="caption">{day.precipitationProbability}%</Typography>
      </Box>

      <Box
        sx={{
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          gap: 0.5,
        }}
      >
        <AirOutlined sx={{ fontSize: 14, color: "text.secondary" }} />
        <Typography variant="caption">
          {Math.round(day.windSpeedMax)} km/h
        </Typography>
      </Box>
    </Paper>
  );
}
