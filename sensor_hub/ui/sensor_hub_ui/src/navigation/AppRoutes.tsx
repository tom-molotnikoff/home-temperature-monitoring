import {BrowserRouter, Route, Routes} from "react-router";
import TemperatureDashboard from "../pages/temperature-dashboard/TemperatureDashboard.tsx";
import SensorsOverview from "../pages/sensors-overview/SensorsOverview.tsx";
import {useSensorContext} from "../hooks/useSensorContext.ts";
import SensorPage from "../pages/sensor/SensorPage.tsx";
import PropertiesOverview from "../pages/properties/PropertiesOverview.tsx";
import LoginPage from "../pages/Login.tsx";
import ChangePasswordPage from "../pages/account/ChangePassword.tsx";
import SessionsPage from "../pages/account/SessionsPage.tsx";
import UsersPage from "../pages/admin/UsersPage.tsx";
import RolesPage from "../pages/admin/RolesPage.tsx";
import OAuthPage from "../pages/admin/OAuthPage.tsx";
import AlertsPage from "../pages/alerts/AlertsPage.tsx";
import NotificationsPage from "../pages/notifications/NotificationsPage.tsx";
import NotificationPreferencesPage from "../pages/notifications/NotificationPreferencesPage.tsx";
import RequireAuth from "./RequireAuth.tsx";


function AppRoutes() {
  const {sensors} = useSensorContext();

  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route path="/account/change-password" element={<RequireAuth><ChangePasswordPage /></RequireAuth>} />
        <Route path="/account/sessions" element={<RequireAuth><SessionsPage /></RequireAuth>} />
        <Route path="/admin/users" element={<RequireAuth><UsersPage /></RequireAuth>} />
        <Route path="/admin/roles" element={<RequireAuth><RolesPage /></RequireAuth>} />
        <Route path="/admin/oauth" element={<RequireAuth><OAuthPage /></RequireAuth>} />
        <Route path="/alerts" element={<RequireAuth><AlertsPage /></RequireAuth>} />
        <Route path="/notifications" element={<RequireAuth><NotificationsPage /></RequireAuth>} />
        <Route path="/notifications/preferences" element={<RequireAuth><NotificationPreferencesPage /></RequireAuth>} />
        <Route path="/" element={<RequireAuth><TemperatureDashboard /></RequireAuth>} />
        <Route path="/sensors-overview" element={<RequireAuth><SensorsOverview /></RequireAuth>} />
        { sensors.map((sensor) => {
          return (
            <Route
              key={sensor.id}
              path={`/sensor/${sensor.id}`}
              element={<RequireAuth><SensorPage sensorId={sensor.id} /></RequireAuth>}
            />
          )
        })}
        <Route path="/properties-overview" element={<RequireAuth><PropertiesOverview /></RequireAuth>} />
      </Routes>
    </BrowserRouter>
  )
}

export default AppRoutes