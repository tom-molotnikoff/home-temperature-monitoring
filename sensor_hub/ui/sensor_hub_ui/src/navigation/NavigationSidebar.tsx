import { Drawer, ListItem, ListItemButton, ListItemText, List, Toolbar, Divider, IconButton, Typography } from '@mui/material';
import {useContext} from "react";
import CloseIcon from '@mui/icons-material/Close';
import {SidebarContext} from "../providers/SidebarContextType.tsx";
import {useNavigate} from "react-router";
import { useAuth } from '../providers/AuthContext.tsx';
import { logout as apiLogout } from '../api/Auth';

function NavigationSidebar() {
  const {open, setOpen} = useContext(SidebarContext);

  const navigate = useNavigate();
  const { user, refresh } = useAuth();

  const pages = [
    { text: 'Sensors', path: '/sensors-overview' },
    { text: 'Temperature', path: '/' },
    { text: 'Properties', path: '/properties-overview' }
  ]

  const handleNavigate = (path: string) => { setOpen(false); navigate(path); };

  const doLogout = async () => {
    try { await apiLogout(); } catch { /* ignore */ }
    await refresh();
    setOpen(false);
    navigate('/login');
  }

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
        {pages.map((page) => (
          <ListItem key={page.text} disablePadding>
            <ListItemButton onClick={() => handleNavigate(page.path)}>
              <ListItemText primary={page.text} />
            </ListItemButton>
          </ListItem>
        ))}
        { user && (
          <>
            <ListItem disablePadding>
              <ListItemButton onClick={() => handleNavigate('/account/sessions')}>
                <ListItemText primary="My sessions" />
              </ListItemButton>
            </ListItem>
            { user.roles?.includes('admin') && (
              <>
                <ListItem disablePadding>
                  <ListItemButton onClick={() => handleNavigate('/admin/users')}>
                    <ListItemText primary="Manage users" />
                  </ListItemButton>
                </ListItem>
                <ListItem disablePadding>
                  <ListItemButton onClick={() => handleNavigate('/admin/roles')}>
                    <ListItemText primary="Manage roles" />
                  </ListItemButton>
                </ListItem>
              </>
            )}
            <ListItem disablePadding>
              <ListItemButton onClick={doLogout}>
                <ListItemText primary="Logout" />
              </ListItemButton>
            </ListItem>
          </>
        )}
        { !user && (
          <ListItem disablePadding>
            <ListItemButton onClick={() => handleNavigate('/login')}>
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