import {DataGrid, type GridColDef, type GridRowParams} from "@mui/x-data-grid";
import type { AlertRule } from "../api/Alerts";

interface AlertRuleDataGridProps {
  handleRowClick?: (params: GridRowParams, event: React.MouseEvent) => void;
  alertRules: AlertRule[]
}

export default function AlertRuleDataGrid({handleRowClick, alertRules}: AlertRuleDataGridProps) {
  const columns: GridColDef[] = [
    { field: 'SensorName', headerName: 'Sensor', flex: 1 },
    { field: 'AlertType', headerName: 'Alert Type', width: 150 },
    { field: 'HighThreshold', headerName: 'High', width: 80 },
    { field: 'LowThreshold', headerName: 'Low', width: 80 },
    { field: 'TriggerStatus', headerName: 'Status', width: 100 },
    { field: 'RateLimitHours', headerName: 'Rate Limit (hrs)', width: 130 },
    { field: 'Enabled', headerName: 'Enabled', width: 80 },
    { field: 'LastAlertSentAt', headerName: 'Last Alert Sent', width: 180 },
  ];

  const rows = alertRules.map(r => ({
    id: r.SensorID,
    ...r,
    HighThreshold: r.HighThreshold ?? '-',
    LowThreshold: r.LowThreshold ?? '-',
    TriggerStatus: r.TriggerStatus || '-',
    Enabled: r.Enabled ? 'Yes' : 'No',
    LastAlertSentAt: r.LastAlertSentAt ? new Date(r.LastAlertSentAt).toLocaleString() : 'Never',
  }));


  return (
    <DataGrid
      rows={rows}
      columns={columns}
      pageSizeOptions={[5, 10, 25]}
      initialState={{ pagination: { paginationModel: { pageSize: 10 } } }}
      onRowClick={handleRowClick}
    />
  )
}