import { Drawer, ListItem, ListItemButton, ListItemText, List, Toolbar, Divider, IconButton, Typography, ListItemIcon } from '@mui/material';
import {useContext} from "react";
import CloseIcon from '@mui/icons-material/Close';
import DeviceThermostatIcon from '@mui/icons-material/DeviceThermostat';
import SensorsIcon from '@mui/icons-material/Sensors';
import SettingsIcon from '@mui/icons-material/Settings';
import HistoryIcon from '@mui/icons-material/History';
import PeopleIcon from '@mui/icons-material/People';
import SecurityIcon from '@mui/icons-material/Security';
import LockOpenIcon from '@mui/icons-material/LockOpen';
import NotificationsActiveIcon from '@mui/icons-material/NotificationsActive';
import NotificationsIcon from '@mui/icons-material/Notifications';
import TuneIcon from '@mui/icons-material/Tune';
import LogoutIcon from '@mui/icons-material/Logout';
import LoginIcon from '@mui/icons-material/Login';
import {SidebarContext} from "../providers/SidebarContextType.tsx";
import {useNavigate} from "react-router";
import { useAuth } from '../providers/AuthContext.tsx';
import { logout as apiLogout } from '../api/Auth';
import {hasPerm} from "../tools/Utils.ts";

function NavigationSidebar() {
  const {open, setOpen} = useContext(SidebarContext);

  const navigate = useNavigate();
  const { user, refresh } = useAuth();

  const handleNavigate = (path: string) => { setOpen(false); navigate(path); };

  const doLogout = async () => {
    try { await apiLogout(); } catch { /* ignore */ }
    await refresh();
    setOpen(false);
    navigate('/login');
  }

  if (user === undefined) return (
    <Drawer
      variant="temporary"
      ModalProps={{
        keepMounted: false,
      }}
      open={open}
      onClose={() => setOpen(false)}
    >
      <Toolbar variant="regular">
        <IconButton edge="start" color="inherit" aria-label="menu" sx={{ mr: 2 }} onClick={() => setOpen(!open)}>
          <CloseIcon />
        </IconButton>
        <Typography variant="h6" color="inherit" component="div">
          Sensor Hub
        </Typography>
      </Toolbar>
      <Divider />
      <List>
        <ListItem disablePadding>
          <ListItemText primary="Loading..." sx={{ padding: 2 }} />
        </ListItem>
      </List>
      <Divider />
    </Drawer>
  );


  return (
    <Drawer
      variant="temporary"
      ModalProps={{
        keepMounted: false,
      }}
      open={open}
      onClose={() => setOpen(false)}
    >
      <Toolbar variant="regular">
        <IconButton edge="start" color="inherit" aria-label="menu" sx={{ mr: 2 }} onClick={() => setOpen(!open)}>
          <CloseIcon />
        </IconButton>
        <Typography variant="h6" color="inherit" component="div">
          Sensor Hub
        </Typography>
      </Toolbar>
      <Divider />
      <List>
        { (hasPerm(user, 'view_readings') && (
          <ListItem disablePadding>
            <ListItemButton onClick={() => handleNavigate('/')}>
              <ListItemIcon><DeviceThermostatIcon /></ListItemIcon>
              <ListItemText primary="Temperature" />
            </ListItemButton>
          </ListItem>
        ))}
        { (hasPerm(user, 'view_sensors') && (
          <ListItem disablePadding>
            <ListItemButton onClick={() => handleNavigate('/sensors-overview')}>
              <ListItemIcon><SensorsIcon /></ListItemIcon>
              <ListItemText primary="Sensors" />
            </ListItemButton>
          </ListItem>
        ))}
        { (hasPerm(user, 'view_properties') && (
          <ListItem disablePadding>
            <ListItemButton onClick={() => handleNavigate('/properties-overview')}>
              <ListItemIcon><SettingsIcon /></ListItemIcon>
              <ListItemText primary="Properties" />
            </ListItemButton>
          </ListItem>
        ))}
        { (hasPerm(user, 'view_alerts') && (
          <ListItem disablePadding>
            <ListItemButton onClick={() => handleNavigate('/alerts')}>
              <ListItemIcon><NotificationsActiveIcon /></ListItemIcon>
              <ListItemText primary="Alerts" />
            </ListItemButton>
          </ListItem>
        ))}
        { (hasPerm(user, 'view_notifications') && (
          <ListItem disablePadding>
            <ListItemButton onClick={() => handleNavigate('/notifications')}>
              <ListItemIcon><NotificationsIcon /></ListItemIcon>
              <ListItemText primary="Notifications" />
            </ListItemButton>
          </ListItem>
        ))}
        { (hasPerm(user, 'manage_notifications') && (
          <ListItem disablePadding>
            <ListItemButton onClick={() => handleNavigate('/notifications/preferences')}>
              <ListItemIcon><TuneIcon /></ListItemIcon>
              <ListItemText primary="Notification Settings" />
            </ListItemButton>
          </ListItem>
        ))}
        { user && (
          <>
            <ListItem disablePadding>
              <ListItemButton onClick={() => handleNavigate('/account/sessions')}>
                <ListItemIcon><HistoryIcon /></ListItemIcon>
                <ListItemText primary="Sessions" />
              </ListItemButton>
            </ListItem>
            { (hasPerm(user,'view_users') || hasPerm(user,'view_roles')) && (
              <>
                { hasPerm(user, 'view_users') && (
                  <ListItem disablePadding>
                    <ListItemButton onClick={() => handleNavigate('/admin/users')}>
                      <ListItemIcon><PeopleIcon /></ListItemIcon>
                      <ListItemText primary="Manage users" />
                    </ListItemButton>
                  </ListItem>
                ) }
                { hasPerm(user, 'view_roles') && (
                  <ListItem disablePadding>
                    <ListItemButton onClick={() => handleNavigate('/admin/roles')}>
                      <ListItemIcon><SecurityIcon /></ListItemIcon>
                      <ListItemText primary="Manage roles" />
                    </ListItemButton>
                  </ListItem>
                ) }
                { hasPerm(user, 'manage_oauth') && (
                  <ListItem disablePadding>
                    <ListItemButton onClick={() => handleNavigate('/admin/oauth')}>
                      <ListItemIcon><LockOpenIcon /></ListItemIcon>
                      <ListItemText primary="OAuth" />
                    </ListItemButton>
                  </ListItem>
                ) }
              </>
            )}
            <ListItem disablePadding>
              <ListItemButton onClick={doLogout}>
                <ListItemIcon><LogoutIcon /></ListItemIcon>
                <ListItemText primary="Logout" />
              </ListItemButton>
            </ListItem>
          </>
        )}
        { !user && (
          <ListItem disablePadding>
            <ListItemButton onClick={() => handleNavigate('/login')}>
              <ListItemIcon><LoginIcon /></ListItemIcon>
              <ListItemText primary="Login" />
            </ListItemButton>
          </ListItem>
        )}
      </List>
      <Divider />
    </Drawer>
  )
}

export default NavigationSidebar