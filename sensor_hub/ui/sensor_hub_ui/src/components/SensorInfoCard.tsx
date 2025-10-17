import type {Sensor} from "../types/types.ts";
import LayoutCard from "../tools/LayoutCard.tsx";
import { CardContent, Chip, Typography, Box, Link, Avatar } from '@mui/material';
import SensorsIcon from '@mui/icons-material/Sensors';

interface SensorInfoCardProps {
  sensor: Sensor
}

function getHealthColor(status: Sensor['healthStatus']) {
  switch (status) {
    case 'good': return 'success';
    case 'bad': return 'error';
    case 'unknown': return 'warning';
    default: return 'default';
  }
}

function getHealthBgColor(status: Sensor['healthStatus']) {
  switch (status) {
    case 'good': return 'success.main';
    case 'bad': return 'error.main';
    case 'unknown': return 'warning.main';
    default: return 'grey.400';
  }
}

function SensorInfoCard({sensor}: SensorInfoCardProps) {
  return (
    <LayoutCard variant="secondary" changes={{alignItems: "center", height: "100%", width: "100%"}}>
      <Box display="flex" alignItems="center" gap={2} mb={2}>
        <Typography variant="h4" height={40}>
          {sensor.name}
        </Typography>
        <Avatar sx={{ bgcolor: getHealthBgColor(sensor.healthStatus), width: 40, height: 40 }}>
          <SensorsIcon />
        </Avatar>
      </Box>
      <CardContent>
        <Box display="flex" flexDirection="column" gap={2}>
          <Box display="flex" alignItems="center" gap={1}>
            <Typography variant="subtitle1">Type:</Typography>
            <Chip label={sensor.type} color="primary" size="small" />
          </Box>
          <Box display="flex" alignItems="center" gap={1}>
            <Typography variant="subtitle1">Health:</Typography>
            <Chip label={sensor.healthStatus} color={getHealthColor(sensor.healthStatus)} size="small" />
          </Box>
          {sensor.healthReason && (
            <Box display="flex" alignItems="center" gap={1}>
              <Typography variant="subtitle1" sx={{
                textWrap: "nowrap"
              }}>Health Reason:</Typography>
              <Typography variant="body2" color="text.secondary">{sensor.healthReason}</Typography>
            </Box>
          )}
          <Box display="flex" alignItems="center" gap={1}>
            <Typography variant="subtitle1">API URL:</Typography>
            <Link href={sensor.url} target="_blank" rel="noopener">{sensor.url}</Link>
          </Box>
        </Box>
      </CardContent>
    </LayoutCard>
  );
}

export default SensorInfoCard;
