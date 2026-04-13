import type {AlertRule} from "../api/Alerts.ts";
import {Box, Chip, Typography} from "@mui/material";

interface AlertRuleCardProps {
  rule: AlertRule;
  onClick: (event: React.MouseEvent) => void;
}

export default function AlertRuleCard({ rule, onClick }: AlertRuleCardProps) {
  return (
    <Box
      onClick={onClick}
      sx={{
        p: 2,
        mb: 1,
        border: '1px solid',
        borderColor: 'divider',
        borderRadius: 1,
        cursor: 'pointer',
        '&:hover': {
          backgroundColor: 'action.hover',
        },
      }}
    >
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={1}>
        <Typography variant="subtitle1" fontWeight="bold">
          {rule.SensorName} — {rule.MeasurementType}
        </Typography>
        <Chip
          label={rule.Enabled ? 'Enabled' : 'Disabled'}
          color={rule.Enabled ? 'success' : 'default'}
          size="small"
        />
      </Box>
      <Typography variant="body2" color="text.secondary">
        {rule.AlertType === 'numeric_range'
          ? `Range: ${rule.LowThreshold ?? '-'} to ${rule.HighThreshold ?? '-'}`
          : `Trigger: ${rule.TriggerStatus || '-'}`
        }
      </Typography>
      {rule.LastAlertSentAt && (
        <Typography variant="caption" color="text.secondary">
          Last alert: {new Date(rule.LastAlertSentAt).toLocaleDateString()}
        </Typography>
      )}
    </Box>
  );
}
