import {BrowserRouter, Route, Routes} from "react-router";
import TemperatureDashboard from "../pages/temperature-dashboard/TemperatureDashboard.tsx";
import SensorsOverview from "../pages/sensors-overview/SensorsOverview.tsx";
import {useSensorContext} from "../hooks/useSensorContext.tsx";
import SensorPage from "../pages/sensor/SensorPage.tsx";


function AppRoutes() {
  const {sensors} = useSensorContext();

  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<TemperatureDashboard />} />
        <Route path="/sensors-overview" element={<SensorsOverview />} />
        { sensors.map((sensor) => {
          return (
            <Route
              key={sensor.id}
              path={`/sensor/${sensor.id}`}
              element={<SensorPage sensorId={sensor.id} />}
            />
          )
        })}
      </Routes>
    </BrowserRouter>
  )
}

export default AppRoutes