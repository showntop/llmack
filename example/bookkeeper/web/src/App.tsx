import { ThemeProvider, createTheme } from '@mui/material/styles';
import { useMediaQuery } from '@mui/material';
import CssBaseline from '@mui/material/CssBaseline';
import { Box, Tab, Tabs, AppBar, Toolbar, Typography, Button, Drawer, IconButton, List, ListItemButton, ListItemIcon, ListItemText } from '@mui/material';
import { useState } from 'react';
import { ChatInterface } from './components/ChatInterface';
import { AnalysisDashboard } from './components/AnalysisDashboard';
import { Auth } from './components/Auth';
import { UserSettings } from './components/UserSettings';
import MenuIcon from '@mui/icons-material/Menu';
import ChatIcon from '@mui/icons-material/Chat';
import BarChartIcon from '@mui/icons-material/BarChart';
import SettingsIcon from '@mui/icons-material/Settings';

const theme = createTheme({
  palette: {
    mode: 'light',
    primary: {
      main: '#1976d2',
    },
  },
  components: {
    MuiDrawer: {
      styleOverrides: {
        paper: {
          width: 240,
        },
      },
    },
  },
});

function App() {
  const [currentTab, setCurrentTab] = useState(0);
  const [token, setToken] = useState<string | null>(localStorage.getItem('token'));
  const [mobileOpen, setMobileOpen] = useState(false);
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));

  const handleTabChange = (event: React.SyntheticEvent, newValue: number) => {
    setCurrentTab(newValue);
    if (isMobile) {
      setMobileOpen(false);
    }
  };

  const handleLogin = (newToken: string) => {
    setToken(newToken);
    localStorage.setItem('token', newToken);
  };

  const handleLogout = () => {
    setToken(null);
    localStorage.removeItem('token');
  };

  const handleDrawerToggle = () => {
    setMobileOpen(!mobileOpen);
  };

  if (!token) {
    return (
      <ThemeProvider theme={theme}>
        <CssBaseline />
        <Auth onLogin={handleLogin} />
      </ThemeProvider>
    );
  }

  const drawer = (
    <Box>
      <List>
        <ListItemButton onClick={() => handleTabChange({} as React.SyntheticEvent, 0)}>
          <ListItemIcon>
            <ChatIcon />
          </ListItemIcon>
          <ListItemText primary="聊天" />
        </ListItemButton>
        <ListItemButton onClick={() => handleTabChange({} as React.SyntheticEvent, 1)}>
          <ListItemIcon>
            <BarChartIcon />
          </ListItemIcon>
          <ListItemText primary="数据分析" />
        </ListItemButton>
        <ListItemButton onClick={() => handleTabChange({} as React.SyntheticEvent, 2)}>
          <ListItemIcon>
            <SettingsIcon />
          </ListItemIcon>
          <ListItemText primary="设置" />
        </ListItemButton>
      </List>
    </Box>
  );

  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <Box sx={{ display: 'flex' }}>
        <AppBar position="fixed" sx={{ zIndex: (theme) => theme.zIndex.drawer + 1 }}>
          <Toolbar>
            {isMobile && (
              <IconButton
                color="inherit"
                aria-label="open drawer"
                edge="start"
                onClick={handleDrawerToggle}
                sx={{ mr: 2 }}
              >
                <MenuIcon />
              </IconButton>
            )}
            <Typography variant="h6" component="div" sx={{ flexGrow: 1 }}>
              智能记账助手
            </Typography>
            <Button color="inherit" onClick={handleLogout}>
              退出
            </Button>
          </Toolbar>
        </AppBar>
        
        {isMobile ? (
          <Drawer
            variant="temporary"
            anchor="left"
            open={mobileOpen}
            onClose={handleDrawerToggle}
            ModalProps={{
              keepMounted: true,
            }}
          >
            {drawer}
          </Drawer>
        ) : (
          <Drawer
            variant="permanent"
            sx={{
              width: 240,
              flexShrink: 0,
              '& .MuiDrawer-paper': {
                width: 240,
                boxSizing: 'border-box',
                marginTop: '64px',
              },
            }}
          >
            {drawer}
          </Drawer>
        )}

        <Box
          component="main"
          sx={{
            flexGrow: 1,
            p: 3,
            width: { sm: `calc(100% - 240px)` },
            marginTop: '64px',
          }}
        >
          {currentTab === 0 && <ChatInterface token={token} />}
          {currentTab === 1 && <AnalysisDashboard token={token} />}
          {currentTab === 2 && <UserSettings token={token} />}
        </Box>
      </Box>
    </ThemeProvider>
  );
}

export default App;
