import React, { useState } from 'react';
import { useNavigate } from 'react-router';
import { login, type LoginResponse } from '../api/Auth';
import { useAuth } from '../providers/AuthContext.tsx';
import {
  Container,
  Box,
  TextField,
  Button,
  Typography,
  Alert,
  CircularProgress,
  Avatar,
  Paper,
} from '@mui/material';
import LockOutlinedIcon from '@mui/icons-material/LockOutlined';

export default function LoginPage() {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();
  const { refresh } = useAuth();

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (loading) return;
    setLoading(true);
    setError(null);
    try {
      const res: LoginResponse = await login(username, password);
      try { await refresh(); } catch { /* ignore */ }
      if (res.must_change_password) {
        navigate('/account/change-password');
      } else {
        navigate('/');
      }
    } catch (err: unknown) {
      let msg = 'Login failed';
      if (err && typeof err === 'object' && 'message' in err) {
        const e = err as { message?: unknown };
        if (typeof e.message === 'string') msg = e.message;
      }
      setError(msg);
    } finally {
      setLoading(false);
    }
  };

  return (
    <Box sx={{
      minHeight: '100vh',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      background: 'linear-gradient(180deg, rgba(245,247,250,1) 0%, rgba(230,235,242,1) 100%)',
      padding: 2
    }}>
      <Container maxWidth="xs">
        <Paper elevation={6} sx={{ p: 4, borderRadius: 2 }}>
          <Box sx={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 2 }}>
            <Avatar sx={{ bgcolor: 'primary.main' }}>
              <LockOutlinedIcon />
            </Avatar>
            <Typography component="h1" variant="h5">Sign in</Typography>
            {error && <Alert severity="error" sx={{ width: '100%' }}>{error}</Alert>}
            <Box component="form" onSubmit={submit} sx={{ mt: 1, width: '100%' }}>
              <TextField
                label="Username"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                fullWidth
                margin="normal"
                autoComplete="username"
                disabled={loading}
                required
              />
              <TextField
                label="Password"
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                fullWidth
                margin="normal"
                autoComplete="current-password"
                disabled={loading}
                required
              />

              <Box sx={{ display: 'flex', justifyContent: 'flex-end', mt: 3 }}>
                <Button
                  type="submit"
                  variant="contained"
                  disabled={loading}
                  startIcon={loading ? <CircularProgress color="inherit" size={18} /> : undefined}
                  sx={{ px: 3 }}
                >
                  {loading ? 'Signing in...' : 'Sign in'}
                </Button>
              </Box>
            </Box>
          </Box>
        </Paper>
      </Container>
    </Box>
  );
}
