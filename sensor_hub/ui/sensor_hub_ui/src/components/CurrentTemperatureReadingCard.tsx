import type { CSSProperties } from "react";
import type { TemperatureReading } from "../types/types";
import LayoutCard from "../tools/LayoutCard.tsx";
import {
  TypographyH3,
  TypographyMutedText,
  TypographySecondaryHeading,
} from "../tools/Typography";

interface CurrentTemperatureReadingCardProps {
  reading: TemperatureReading;
}

function CurrentTemperatureReadingCard({
  reading,
}: CurrentTemperatureReadingCardProps) {
  return (
    <LayoutCard
      changes={layoutCardStyleChanges}
      testid="current-temperature-card"
      variant="tertiary"
    >
      <TypographyH3>{reading.sensor_name}</TypographyH3>
      <TypographySecondaryHeading>
        {reading.temperature ?? "N/A"}°C
      </TypographySecondaryHeading>
      <TypographyMutedText>
        {reading.time
          ? new Date(
              reading.time.replace(" ", "T")
            ).toLocaleTimeString()
          : "Unknown time"}
      </TypographyMutedText>
    </LayoutCard>
  );
}

const layoutCardStyleChanges: CSSProperties = {
  borderRadius: 8,
  padding: "18px 24px",
  minWidth: 190,
  alignItems: "center",
};

export default CurrentTemperatureReadingCard;
