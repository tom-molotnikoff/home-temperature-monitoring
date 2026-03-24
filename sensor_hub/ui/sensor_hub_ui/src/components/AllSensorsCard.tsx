import SensorsDataGrid from './SensorsDataGrid';
import { useSensorContext } from '../hooks/useSensorContext';
import { useAuth } from '../providers/AuthContext';

export default function AllSensorsCard() {
  const { sensors } = useSensorContext();
  const { user } = useAuth();

  if (!user) return null;

  return (
    <SensorsDataGrid
      cardHeight="500px"
      sensors={sensors}
      showReason={true}
      showType={true}
      showEnabled={true}
      user={user}
    />
  );
}
