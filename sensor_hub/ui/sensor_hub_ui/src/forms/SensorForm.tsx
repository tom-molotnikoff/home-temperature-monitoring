import type {Sensor} from "../types/types.ts";

interface SensorFormProps {
  sensor?: Sensor
}

function SensorForm ({ sensor } : SensorFormProps) {
  if (!sensor) {
    return <div>No sensor</div>;
  }
  return (<><h1>{sensor.name}</h1></>);
}

export default SensorForm;