import type { CSSProperties } from "@mui/material";
import { useCurrentTemperatures } from "../hooks/useCurrentTemperatures";
import CurrentTemperatureReadingCard from "./CurrentTemperatureReadingCard";
import LoadingContentBlock from "../tools/LoadingContentBlock";
import { TypographyH2 } from "../tools/Typography";
import CenteredFlex from "../tools/CenteredFlex";
import DesktopRowMobileColumn from "../tools/DesktopRowMobileColumn";

function CurrentTemperatures() {
  const currentTemperatures = useCurrentTemperatures();

  const sensorNames = Object.keys(currentTemperatures).sort((a, b) =>
    a.localeCompare(b)
  );

  return (
    <CenteredFlex>
      <TypographyH2>Current Temperatures</TypographyH2>
      <DesktopRowMobileColumn>
        {sensorNames.map((sensor) => {
          const readingObj = currentTemperatures[sensor];
          return (
            <CurrentTemperatureReadingCard key={sensor} reading={readingObj} />
          );
        })}
        {sensorNames.length === 0 && (
          <LoadingContentBlock changes={loadingStyleChanges}>
            Checking Temperatures...
          </LoadingContentBlock>
        )}
      </DesktopRowMobileColumn>
    </CenteredFlex>
  );
}

const loadingStyleChanges: CSSProperties = {
  minWidth: 395,
  minHeight: 142,
};

export default CurrentTemperatures;
