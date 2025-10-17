import {AppBar, IconButton, Menu, MenuItem, Toolbar, Typography, useColorScheme} from '@mui/material';
import MenuIcon from '@mui/icons-material/Menu';
import WbSunnyIcon from '@mui/icons-material/WbSunny';
import DarkModeIcon from '@mui/icons-material/DarkMode';
import LaptopIcon from '@mui/icons-material/Laptop';
import CheckIcon from '@mui/icons-material/Check';
import {SidebarContext} from "../providers/SidebarContextType.tsx";
import {useContext, useState} from "react";
import {useIsMobile} from "../hooks/useMobile.ts";

interface TopAppBarProps {
  pageTitle: string;
}

function TopAppBar({ pageTitle }: TopAppBarProps) {
  const {open, setOpen} = useContext(SidebarContext);
  const {mode, setMode} = useColorScheme();
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
  const openMenu = Boolean(anchorEl);
  const isMobile = useIsMobile();

  const handleMenuOpen = (event: React.MouseEvent<HTMLElement>) => {
    setAnchorEl(event.currentTarget);
  };

  const handleMenuClose = () => {
    setAnchorEl(null);
  };

  const handleModeChange = (newMode: 'light' | 'dark' | 'system') => {
    setMode(newMode);
    console.log(`Theme mode changed to: ${newMode}`);
    handleMenuClose();
  };

  let ModeIcon = WbSunnyIcon;
  if (mode === 'dark') ModeIcon = DarkModeIcon;
  else if (mode === 'system') ModeIcon = LaptopIcon;

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
      </Toolbar>
    </AppBar>
  )
}

export default TopAppBar;