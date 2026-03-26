import { TypographyH2 } from "../tools/Typography.tsx";
import LayoutCard from "../tools/LayoutCard.tsx";
import { useSensorContext } from "../hooks/useSensorContext.ts";
import SensorTypePieChart from "./SensorTypePieChart.tsx";
import { CircularProgress, Box } from "@mui/material";
import EmptyState from "./EmptyState";
import CategoryOutlinedIcon from "@mui/icons-material/CategoryOutlined";

function SensorTypeCard() {
  const { sensors, loaded } = useSensorContext();

  return (
    <LayoutCard
      variant="secondary"
      changes={{ alignItems: "center", height: "100%", width: "100%" }}
    >
      <TypographyH2>Sensor Types</TypographyH2>
      {!loaded ? (
        <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: 200 }}>
          <CircularProgress />
        </Box>
      ) : sensors.length === 0 ? (
        <EmptyState
          icon={<CategoryOutlinedIcon sx={{ fontSize: 48 }} />}
          title="No sensors to categorise"
          description="Sensor type breakdown will appear here once sensors are added."
        />
      ) : (
        <SensorTypePieChart sensors={sensors} />
      )}
    </LayoutCard>
  );
}

export default SensorTypeCard;
