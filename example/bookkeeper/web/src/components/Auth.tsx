import React, { useState } from 'react';
import {
  Box,
  Button,
  TextField,
  Typography,
  Paper,
  Tab,
  Tabs,
  Alert,
} from '@mui/material';
import axios from 'axios';

interface AuthProps {
  onLogin: (token: string) => void;
}

export const Auth: React.FC<AuthProps> = ({ onLogin }) => {
  const [tab, setTab] = useState(0);
  const [username, setUsername] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');

  const handleTabChange = (event: React.SyntheticEvent, newValue: number) => {
    setTab(newValue);
    setError('');
  };

  const handleRegister = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      const response = await axios.post('/api/auth/register', {
        username,
        email,
        password,
      });
      setError('');
      setTab(0); // 切换到登录标签
    } catch (err: any) {
      setError(err.response?.data?.error || '注册失败');
    }
  };

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      const response = await axios.post('/api/auth/login', {
        username,
        password,
      });
      onLogin(response.data.token);
      setError('');
    } catch (err: any) {
      setError(err.response?.data?.error || '登录失败');
    }
  };

  return (
    <Box
      display="flex"
      justifyContent="center"
      alignItems="center"
      minHeight="100vh"
      bgcolor="background.default"
    >
      <Paper elevation={3} sx={{ p: 4, width: '100%', maxWidth: 400 }}>
        <Typography variant="h5" component="h1" gutterBottom align="center">
          智能记账助手
        </Typography>

        <Tabs value={tab} onChange={handleTabChange} centered sx={{ mb: 3 }}>
          <Tab label="登录" />
          <Tab label="注册" />
        </Tabs>

        {error && (
          <Alert severity="error" sx={{ mb: 2 }}>
            {error}
          </Alert>
        )}

        <form onSubmit={tab === 0 ? handleLogin : handleRegister}>
          <TextField
            fullWidth
            label="用户名"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            margin="normal"
            required
          />

          {tab === 1 && (
            <TextField
              fullWidth
              label="邮箱"
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              margin="normal"
              required
            />
          )}

          <TextField
            fullWidth
            label="密码"
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            margin="normal"
            required
          />

          <Button
            fullWidth
            type="submit"
            variant="contained"
            color="primary"
            sx={{ mt: 3 }}
          >
            {tab === 0 ? '登录' : '注册'}
          </Button>
        </form>
      </Paper>
    </Box>
  );
}; 