import { Box, Typography, Button } from '@mui/material';
import { useNavigate } from 'react-router';

interface EmptyStateProps {
  icon?: React.ReactNode;
  title: string;
  description?: string;
  actionLabel?: string;
  actionHref?: string;
  onAction?: () => void;
  minHeight?: number | string;
}

export default function EmptyState({
  icon,
  title,
  description,
  actionLabel,
  actionHref,
  onAction,
  minHeight = 200,
}: EmptyStateProps) {
  const navigate = useNavigate();

  const handleClick = () => {
    if (onAction) {
      onAction();
    } else if (actionHref) {
      navigate(actionHref);
    }
  };

  return (
    <Box
      sx={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        minHeight,
        gap: 1.5,
        py: 4,
        px: 2,
        textAlign: 'center',
      }}
    >
      <Box sx={{ color: 'text.disabled' }}>{icon}</Box>
      <Typography variant="body1" fontWeight={600}>
        {title}
      </Typography>
      {description && (
        <Typography variant="body2" color="text.secondary" sx={{ maxWidth: 360 }}>
          {description}
        </Typography>
      )}
      {actionLabel && (onAction || actionHref) && (
        <Button variant="outlined" size="small" onClick={handleClick} sx={{ mt: 1 }}>
          {actionLabel}
        </Button>
      )}
    </Box>
  );
}
