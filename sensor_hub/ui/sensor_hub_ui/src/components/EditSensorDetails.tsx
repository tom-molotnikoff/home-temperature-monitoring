import type {Sensor} from "../types/types.ts";
import SensorForm from "../forms/SensorForm.tsx";
import LayoutCard from "../tools/LayoutCard.tsx";


interface EditSensorDetailsProps {
  sensor: Sensor
}


function EditSensorDetails({ sensor } : EditSensorDetailsProps) {

  return (
    <LayoutCard variant={"secondary"} changes={{alignItems: "center", height: "100%", width: "100%"}}>
      <SensorForm sensor={sensor} />
    </LayoutCard>
  )
}

export default EditSensorDetails;