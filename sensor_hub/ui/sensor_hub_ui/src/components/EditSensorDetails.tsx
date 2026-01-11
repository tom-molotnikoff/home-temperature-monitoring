import type {Sensor} from "../types/types.ts";
import SensorForm from "../forms/SensorForm.tsx";
import LayoutCard from "../tools/LayoutCard.tsx";
import {useAuth} from "../providers/AuthContext.tsx";


interface EditSensorDetailsProps {
  sensor: Sensor
}


function EditSensorDetails({ sensor } : EditSensorDetailsProps) {
  const { user } = useAuth();
  if (user === undefined) {
    return (
      <LayoutCard variant={"secondary"} changes={{alignItems: "center", height: "100%", width: "100%"}}>
        Loading...
      </LayoutCard>
    )
  }

  return (
    <LayoutCard variant={"secondary"} changes={{alignItems: "center", height: "100%", width: "100%"}}>
      <SensorForm sensor={sensor} mode="edit" user={user} />
    </LayoutCard>
  )
}

export default EditSensorDetails;