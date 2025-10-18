import SensorForm from "../forms/SensorForm.tsx";
import LayoutCard from "../tools/LayoutCard.tsx";

function AddNewSensor() {

  return (
    <LayoutCard variant={"secondary"} changes={{alignItems: "center", height: "100%", width: "100%"}}>
      <SensorForm mode="create" />
    </LayoutCard>
  )
}

export default AddNewSensor;