import React, { useState } from 'react';
import { changePassword } from '../../api/Users';
import { useNavigate } from 'react-router';
import { Container, Box, TextField, Button, Typography, Alert, Avatar, Paper, CircularProgress } from '@mui/material';
import LockOutlinedIcon from '@mui/icons-material/LockOutlined';
import PageContainer from '../../tools/PageContainer';

function extractErrorMessage(err: unknown): string | null {
  if (!err) return null;
  if (typeof err === 'string') return err;
  if (typeof err === 'object' && err !== null && 'message' in err) {
    const e = err as { message?: unknown };
    if (typeof e.message === 'string') return e.message;
    return String(e.message);
  }
  return String(err);
}

export default function ChangePasswordPage() {
  const [newPassword, setNewPassword] = useState('');
  const [confirm, setConfirm] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    if (newPassword !== confirm) {
      setError('Passwords do not match');
      return;
    }
    setLoading(true);
    try {
      await changePassword(0, newPassword);
      navigate('/');
    } catch (err: unknown) {
      const message = extractErrorMessage(err) || 'Failed to change password';
      setError(message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <PageContainer titleText="Change password">
      <Container maxWidth="sm">
        <Paper elevation={4} sx={{ p: 4, borderRadius: 2, mt: 4 }}>
          <Box sx={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 2 }}>
            <Avatar sx={{ bgcolor: 'primary.main' }}>
              <LockOutlinedIcon />
            </Avatar>
            <Typography component="h1" variant="h5">Change password</Typography>
            {error && <Alert severity="error" sx={{ width: '100%' }}>{error}</Alert>}
            <Box component="form" onSubmit={submit} sx={{ mt: 1, width: '100%' }}>
              <TextField label="New password" type="password" value={newPassword} onChange={(e) => setNewPassword(e.target.value)} fullWidth margin="normal" disabled={loading} required />
              <TextField label="Confirm password" type="password" value={confirm} onChange={(e) => setConfirm(e.target.value)} fullWidth margin="normal" disabled={loading} required />

              <Box sx={{ display: 'flex', justifyContent: 'flex-end', mt: 2 }}>
                <Button type="submit" variant="contained" disabled={loading} startIcon={loading ? <CircularProgress color="inherit" size={18} /> : undefined}>
                  {loading ? 'Saving...' : 'Change password'}
                </Button>
              </Box>
            </Box>
          </Box>
        </Paper>
      </Container>
    </PageContainer>
  );
}
