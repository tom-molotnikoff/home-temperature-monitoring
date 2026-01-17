import {AppBar, IconButton, Menu, MenuItem, Toolbar, Typography, useColorScheme, Avatar, ListItemIcon} from '@mui/material';
import MenuIcon from '@mui/icons-material/Menu';
import WbSunnyIcon from '@mui/icons-material/WbSunny';
import DarkModeIcon from '@mui/icons-material/DarkMode';
import LaptopIcon from '@mui/icons-material/Laptop';
import HistoryIcon from '@mui/icons-material/History';
import CheckIcon from '@mui/icons-material/Check';
import AccountCircle from '@mui/icons-material/AccountCircle';
import ExitToAppIcon from '@mui/icons-material/ExitToApp';
import PeopleIcon from '@mui/icons-material/People';
import {SidebarContext} from "../providers/SidebarContextType.tsx";
import {useContext, useState} from "react";
import {useIsMobile} from "../hooks/useMobile.ts";
import { useNavigate } from 'react-router';
import { useAuth } from '../providers/AuthContext.tsx';
import { logout as apiLogout } from '../api/Auth';
import {hasPerm} from "../tools/Utils.ts";
import SecurityIcon from "@mui/icons-material/Security";
import NotificationBell from "../components/NotificationBell";

interface TopAppBarProps {
  pageTitle: string;
}

function TopAppBar({ pageTitle }: TopAppBarProps) {
  const {open, setOpen} = useContext(SidebarContext);
  const {mode, setMode} = useColorScheme();
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
  const [accountAnchor, setAccountAnchor] = useState<null | HTMLElement>(null);
  const openMenu = Boolean(anchorEl);
  const openAccount = Boolean(accountAnchor);
  const isMobile = useIsMobile();
  const navigate = useNavigate();
  const { user, refresh } = useAuth();

  const handleMenuOpen = (event: React.MouseEvent<HTMLElement>) => {
    setAnchorEl(event.currentTarget);
  };

  const handleMenuClose = () => {
    setAnchorEl(null);
  };

  const handleAccountOpen = (event: React.MouseEvent<HTMLElement>) => {
    setAccountAnchor(event.currentTarget);
  };
  const handleAccountClose = () => setAccountAnchor(null);

  const handleModeChange = (newMode: 'light' | 'dark' | 'system') => {
    setMode(newMode);
    handleMenuClose();
  };

  const doLogout = async () => {
    try {
      await apiLogout();
    } catch {
      // ignore
    }
    await refresh();
    handleAccountClose();
    navigate('/login');
  };

  let ModeIcon = WbSunnyIcon;
  if (mode === 'dark') ModeIcon = DarkModeIcon;
  else if (mode === 'system') ModeIcon = LaptopIcon;

  const accountMenuItems: React.ReactNode[] = [];
  if (user) {
    accountMenuItems.push(
      <MenuItem key="mysessions" onClick={() => { handleAccountClose(); navigate('/account/sessions'); }}>
        <ListItemIcon><HistoryIcon fontSize="small" /></ListItemIcon>
        My sessions
      </MenuItem>
    );
    accountMenuItems.push(
      <MenuItem key="changepw" onClick={() => { handleAccountClose(); navigate('/account/change-password'); }}>
        <ListItemIcon><AccountCircle fontSize="small" /></ListItemIcon>
        Change password
      </MenuItem>
    );
    if (hasPerm(user, "manage_users")) {
      accountMenuItems.push(
        <MenuItem key="manageusers" onClick={() => { handleAccountClose(); navigate('/admin/users'); }}>
          <ListItemIcon><PeopleIcon fontSize="small" /></ListItemIcon>
          Manage users
        </MenuItem>
      );
    }
    if (hasPerm(user, "manage_roles")) {
      accountMenuItems.push(
        <MenuItem key="manageroles" onClick={() => { handleAccountClose(); navigate('/admin/roles'); }}>
          <ListItemIcon><SecurityIcon fontSize="small" /></ListItemIcon>
          Manage roles
        </MenuItem>
      );
    }
    accountMenuItems.push(
      <MenuItem key="logout" onClick={doLogout}>
        <ListItemIcon><ExitToAppIcon fontSize="small" /></ListItemIcon>
        Logout
      </MenuItem>
    );
  } else {
    accountMenuItems.push(
      <MenuItem key="login" onClick={() => { handleAccountClose(); navigate('/login'); }}>
        <ListItemIcon><AccountCircle fontSize="small" /></ListItemIcon>
        Login
      </MenuItem>
    );
  }

  return (
    <AppBar position="sticky">
      <Toolbar variant="regular">
        <IconButton edge="start" color="inherit" aria-label="menu" sx={{ mr: 2 }} onClick={() => setOpen(!open)}>
          <MenuIcon />
        </IconButton>
        {isMobile ? null : (<Typography variant="h6" color="inherit" component="div" sx={{minWidth: "fit-content"}}>
          Sensor Hub
        </Typography>)}

        <Typography variant="h6" color="inherit" component="div" sx={{ flexGrow: 1, textAlign: 'end', minWidth: "fit-content" }}>
          {pageTitle}
        </Typography>
        {user && hasPerm(user, 'view_notifications') && <NotificationBell />}
        <IconButton
          color="inherit"
          aria-label="theme switcher"
          onClick={handleMenuOpen}
          sx={{ ml: 2 }}
        >
          <ModeIcon />
        </IconButton>
        <Menu
          anchorEl={anchorEl}
          open={openMenu}
          onClose={handleMenuClose}
          anchorOrigin={{ vertical: 'bottom', horizontal: 'right' }}
          transformOrigin={{ vertical: 'top', horizontal: 'right' }}
        >
          <MenuItem selected={mode === 'light'} onClick={() => handleModeChange('light')}>
            Light
            {mode === 'light' && <CheckIcon fontSize="small" sx={{ ml: 2 }} />}
          </MenuItem>
          <MenuItem selected={mode === 'dark'} onClick={() => handleModeChange('dark')}>
            Dark
            {mode === 'dark' && <CheckIcon fontSize="small" sx={{ ml: 2 }} />}
          </MenuItem>
          <MenuItem selected={mode === 'system'} onClick={() => handleModeChange('system')}>
            System
            {mode === 'system' && <CheckIcon fontSize="small" sx={{ ml: 2 }} />}
          </MenuItem>
        </Menu>

        <IconButton color="inherit" onClick={handleAccountOpen} sx={{ ml: 1 }}>
          <Avatar sx={{ width: 32, height: 32 }}>{user?.username?.charAt(0).toUpperCase() ?? 'S'}</Avatar>
        </IconButton>
        <Menu anchorEl={accountAnchor} open={openAccount} onClose={handleAccountClose} anchorOrigin={{ vertical: 'bottom', horizontal: 'right' }} transformOrigin={{ vertical: 'top', horizontal: 'right' }}>
          {accountMenuItems}
        </Menu>
      </Toolbar>
    </AppBar>
  )
}

export default TopAppBar;
