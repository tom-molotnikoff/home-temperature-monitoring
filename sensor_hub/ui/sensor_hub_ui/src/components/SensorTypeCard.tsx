import { TypographyH2 } from "../tools/Typography.tsx";
import LayoutCard from "../tools/LayoutCard.tsx";
import { useSensorContext } from "../hooks/useSensorContext.ts";
import SensorTypePieChart from "./SensorTypePieChart.tsx";

function SensorTypeCard() {
  const { sensors } = useSensorContext();

  return (
    <LayoutCard
      variant="secondary"
      changes={{ alignItems: "center", height: "100%", width: "100%" }}
    >
      <TypographyH2>Sensor Types</TypographyH2>
      <SensorTypePieChart sensors={sensors} />
    </LayoutCard>
  );
}

export default SensorTypeCard;
