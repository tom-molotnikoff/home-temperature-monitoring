import type {Sensor} from "../types/types.ts";
import LayoutCard from "../tools/LayoutCard.tsx";
import {TypographyH2, TypographyH3, TypographyMutedText} from "../tools/Typography.tsx";

interface SensorSummaryCardProps {
  sensors: Sensor[]
}

function SensorSummaryCard({ sensors }: SensorSummaryCardProps) {

  return (
    <LayoutCard variant="secondary">
      <TypographyH2>Sensor Summary</TypographyH2>
      <TypographyH3>Total Sensors: {sensors.length}</TypographyH3>
      <LayoutCard variant="primary" changes={{padding: 0, border: "none"}} direction="column">
        {sensors.map(sensor => (
          <LayoutCard
            variant="tertiary"
            changes={{minWidth: "fit-content", flex: 1}}
          >
            <TypographyH3>Sensor Name: {sensor.name}</TypographyH3>
            <TypographyMutedText>
              API URL: {sensor.url}
            </TypographyMutedText>
          </LayoutCard>
        ))}
      </LayoutCard>
    </LayoutCard>
  )
}



export default SensorSummaryCard