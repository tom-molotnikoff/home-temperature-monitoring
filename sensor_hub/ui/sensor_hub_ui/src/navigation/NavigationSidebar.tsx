import { Drawer, ListItem, ListItemButton, ListItemText, List, Toolbar, Divider, IconButton, Typography } from '@mui/material';
import {useContext} from "react";
import CloseIcon from '@mui/icons-material/Close';
import {SidebarContext} from "../providers/SidebarContextType.tsx";
import {useNavigate} from "react-router";

function NavigationSidebar() {
  const {open, setOpen} = useContext(SidebarContext);

  const navigate = useNavigate();

  const pages = [
    { text: 'Sensors', path: '/sensors-overview' },
    { text: 'Temperature', path: '/' },
    { text: 'Properties', path: '/properties-overview' }
  ]

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
            <ListItemButton onClick={() => {setOpen(false); navigate(page.path)}}>
              <ListItemText primary={page.text} />
            </ListItemButton>
          </ListItem>
        ))}
      </List>
      <Divider />
    </Drawer>
  )
}

export default NavigationSidebar