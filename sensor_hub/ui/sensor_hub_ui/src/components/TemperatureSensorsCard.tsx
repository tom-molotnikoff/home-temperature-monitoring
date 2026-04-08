import SensorsDataGrid from './SensorsDataGrid';
import { useSensorContext } from '../hooks/useSensorContext';
import { useAuth } from '../providers/AuthContext';

export default function TemperatureSensorsCard() {
  const { sensors } = useSensorContext();
  const { user } = useAuth();

  if (!user) return null;

  const temperatureSensors = sensors.filter(s => s.sensorDriver === 'sensor-hub-http-temperature');

  return (
    <SensorsDataGrid
      sensors={temperatureSensors}
      cardHeight="100%"
      showReason={false}
      showType={false}
      showEnabled={true}
      title="Temperature Sensors"
      user={user}
    />
  );
}
