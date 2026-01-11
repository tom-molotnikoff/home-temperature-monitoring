import SensorForm from "../forms/SensorForm.tsx";
import LayoutCard from "../tools/LayoutCard.tsx";
import {useAuth} from "../providers/AuthContext.tsx";

function AddNewSensor() {
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
      <SensorForm mode="create" user={ user }/>
    </LayoutCard>
  )
}

export default AddNewSensor;