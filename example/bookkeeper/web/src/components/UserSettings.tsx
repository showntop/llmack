import React, { useState, useEffect } from 'react';
import {
  Box,
  Card,
  CardContent,
  Typography,
  TextField,
  Button,
  Avatar,
  IconButton,
  Divider,
  Alert,
  CircularProgress,
  useTheme,
  useMediaQuery,
  Stack,
} from '@mui/material';
import PhotoCamera from '@mui/icons-material/PhotoCamera';
import axios from 'axios';

interface UserSettingsProps {
  token: string;
}

interface UserProfile {
  id: number;
  username: string;
  email: string;
  avatar?: string;
}

export const UserSettings: React.FC<UserSettingsProps> = ({ token }) => {
  const [profile, setProfile] = useState<UserProfile | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');

  // 个人信息表单
  const [username, setUsername] = useState('');
  const [email, setEmail] = useState('');

  // 密码表单
  const [oldPassword, setOldPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');

  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));

  useEffect(() => {
    fetchProfile();
  }, [token]);

  const fetchProfile = async () => {
    try {
      const response = await axios.get('http://localhost:8080/api/user/profile', {
        headers: {
          Authorization: token,
        },
      });
      setProfile(response.data.user);
      setUsername(response.data.user.username);
      setEmail(response.data.user.email);
    } catch (err: any) {
      setError(err.response?.data?.error || '获取用户信息失败');
    } finally {
      setLoading(false);
    }
  };

  const handleUpdateProfile = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setSuccess('');

    try {
      await axios.put(
        'http://localhost:8080/api/user/profile',
        {
          username,
          email,
        },
        {
          headers: {
            Authorization: token,
          },
        }
      );
      setSuccess('个人信息更新成功');
    } catch (err: any) {
      setError(err.response?.data?.error || '更新个人信息失败');
    }
  };

  const handleUpdatePassword = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setSuccess('');

    if (newPassword !== confirmPassword) {
      setError('两次输入的密码不一致');
      return;
    }

    try {
      await axios.put(
        'http://localhost:8080/api/user/password',
        {
          old_password: oldPassword,
          new_password: newPassword,
        },
        {
          headers: {
            Authorization: token,
          },
        }
      );
      setSuccess('密码更新成功');
      setOldPassword('');
      setNewPassword('');
      setConfirmPassword('');
    } catch (err: any) {
      setError(err.response?.data?.error || '更新密码失败');
    }
  };

  const handleAvatarUpload = async (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (!file) return;

    const formData = new FormData();
    formData.append('avatar', file);

    try {
      await axios.post('http://localhost:8080/api/user/avatar', formData, {
        headers: {
          Authorization: token,
          'Content-Type': 'multipart/form-data',
        },
      });
      setSuccess('头像更新成功');
      fetchProfile();
    } catch (err: any) {
      setError(err.response?.data?.error || '更新头像失败');
    }
  };

  if (loading) {
    return (
      <Box display="flex" justifyContent="center" alignItems="center" minHeight="400px">
        <CircularProgress />
      </Box>
    );
  }

  return (
    <Box p={isMobile ? 1 : 3}>
      {error && (
        <Alert severity="error" sx={{ mb: 2 }}>
          {error}
        </Alert>
      )}
      {success && (
        <Alert severity="success" sx={{ mb: 2 }}>
          {success}
        </Alert>
      )}

      <Stack spacing={isMobile ? 2 : 3}>
        <Card>
          <CardContent>
            <Typography variant="h6" gutterBottom>
              个人信息
            </Typography>
            <Box sx={{ display: 'flex', alignItems: 'center', mb: 3, flexDirection: isMobile ? 'column' : 'row' }}>
              <Avatar
                src={profile?.avatar}
                sx={{ width: isMobile ? 80 : 100, height: isMobile ? 80 : 100, mr: isMobile ? 0 : 2, mb: isMobile ? 2 : 0 }}
              />
              <Box sx={{ textAlign: isMobile ? 'center' : 'left' }}>
                <input
                  accept="image/*"
                  style={{ display: 'none' }}
                  id="avatar-upload"
                  type="file"
                  onChange={handleAvatarUpload}
                />
                <label htmlFor="avatar-upload">
                  <IconButton component="span" color="primary">
                    <PhotoCamera />
                  </IconButton>
                </label>
                <Typography variant="body2" color="text.secondary">
                  点击上传新头像
                </Typography>
              </Box>
            </Box>

            <form onSubmit={handleUpdateProfile}>
              <TextField
                fullWidth
                label="用户名"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                margin="normal"
                required
                size={isMobile ? "small" : "medium"}
              />
              <TextField
                fullWidth
                label="邮箱"
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                margin="normal"
                required
                size={isMobile ? "small" : "medium"}
              />
              <Button
                type="submit"
                variant="contained"
                color="primary"
                sx={{ mt: 2 }}
                fullWidth={isMobile}
              >
                保存修改
              </Button>
            </form>
          </CardContent>
        </Card>

        <Card>
          <CardContent>
            <Typography variant="h6" gutterBottom>
              修改密码
            </Typography>
            <form onSubmit={handleUpdatePassword}>
              <TextField
                fullWidth
                label="当前密码"
                type="password"
                value={oldPassword}
                onChange={(e) => setOldPassword(e.target.value)}
                margin="normal"
                required
                size={isMobile ? "small" : "medium"}
              />
              <TextField
                fullWidth
                label="新密码"
                type="password"
                value={newPassword}
                onChange={(e) => setNewPassword(e.target.value)}
                margin="normal"
                required
                size={isMobile ? "small" : "medium"}
              />
              <TextField
                fullWidth
                label="确认新密码"
                type="password"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                margin="normal"
                required
                size={isMobile ? "small" : "medium"}
              />
              <Button
                type="submit"
                variant="contained"
                color="primary"
                sx={{ mt: 2 }}
                fullWidth={isMobile}
              >
                更新密码
              </Button>
            </form>
          </CardContent>
        </Card>
      </Stack>
    </Box>
  );
}; 