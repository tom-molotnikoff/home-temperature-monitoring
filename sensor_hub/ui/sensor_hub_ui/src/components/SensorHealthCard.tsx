import {TypographyH2} from "../tools/Typography.tsx";
import SensorHealthPieChart from "./SensorHealthPieChart.tsx";
import LayoutCard from "../tools/LayoutCard.tsx";
import {useSensorContext} from "../hooks/useSensorContext.tsx";


function SensorHealthCard() {
  const {sensors } = useSensorContext();

  return (
    <LayoutCard variant="secondary" changes={{alignItems: "center", height: "100%", width: "100%"}}>
      <TypographyH2>Sensor Health</TypographyH2>
      <SensorHealthPieChart sensors={sensors}/>
    </LayoutCard>
  )
}

export default SensorHealthCard;