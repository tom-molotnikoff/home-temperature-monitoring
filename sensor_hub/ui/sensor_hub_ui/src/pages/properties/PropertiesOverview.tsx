import {Box, Grid, List, ListItem, Paper, Typography, Button, TextField, Stack, CircularProgress, Snackbar, Alert} from "@mui/material";
import PageContainer from "../../tools/PageContainer.tsx";
import {useEffect, useState} from "react";
import { PropertiesApi } from "../../api/Properties.ts";
import type { PropertiesApiStructure } from "../../types/types.ts";
import { useIsMobile } from "../../hooks/useMobile";
import {useProperties} from "../../hooks/useProperties.ts";


export default function PropertiesOverview() {
  const properties = useProperties();
  const [editedProperties, setEditedProperties] = useState<PropertiesApiStructure | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);
  const [saved, setSaved] = useState(false);
  const isMobile = useIsMobile();

  useEffect(() => {
    if (properties) {
      setEditedProperties(properties);
    }
  }, [properties]);

  const handleChange = (key: string, value: string) => {
    setEditedProperties(prev => {
      if (!prev) return prev;
      return { ...prev, [key]: value };
    });
  }

  const isDirty = () => {
    if (!properties && !editedProperties) return false;
    if (!properties || !editedProperties) return true;
    return JSON.stringify(properties) !== JSON.stringify(editedProperties);
  }

  const handleSave = async () => {
    if (!editedProperties) return;
    setSaving(true);
    setError(null);
    setSaved(false);
    try {
      await PropertiesApi.patchProperties(editedProperties);
      setSaved(true);
    } catch (e: unknown) {
      console.error("Failed to save properties", e);
      let msg: string;
      if (e && typeof e === 'object' && 'message' in e && typeof (e as {message?: unknown}).message === 'string') {
        msg = (e as {message: string}).message;
      } else {
        try {
          msg = JSON.stringify(e);
        } catch {
          msg = String(e);
        }
      }
      setError(msg);
    } finally {
      setSaving(false);
      setTimeout(() => setSaved(false), 2000);
    }
  }

  return (
    <PageContainer titleText="Properties Overview">
      <Box sx={{flexGrow: 1, width: '100%'}}>
        <Grid
            container
            spacing={2}
            alignItems="stretch"
            sx={{ minHeight: "100%" }}
        >
          <Grid size={12}>
            {editedProperties && (
              <Paper sx={{ padding: 2, width: '100%', maxWidth: '100%', alignSelf: 'stretch' }}>
                <Stack direction="row" justifyContent="space-between" alignItems="center" sx={{ mb: 2 }}>
                  <Typography variant="h4">Properties</Typography>
                  <Stack direction="row" spacing={1} alignItems="center">
                    {saving && <CircularProgress size={20} />}
                    <Button variant="contained" color="primary" onClick={handleSave} disabled={!isDirty() || saving}>
                      Save changes
                    </Button>
                  </Stack>
                </Stack>

                {error && (
                  <Typography color="error">Error: {error}</Typography>
                )}

                {!properties && !error && (
                  <Typography>Loading properties...</Typography>
                )}

                <List sx={{ width: '100%' }}>
                  {Object.entries(editedProperties).map(([key, value]) => (
                    <ListItem key={key} divider sx={{ display: 'flex', flexDirection: isMobile ? 'column' : 'row', alignItems: isMobile ? 'stretch' : 'center', width: '100%', py: 1, gap: 2 }}>
                      <Box sx={{ flex: isMobile ? '0 0 auto' : '0 0 420px', minWidth: 160, pr: isMobile ? 0 : 2, wordBreak: 'break-all', width: isMobile ? '100%' : undefined }}>
                        <Typography sx={{ fontFamily: 'monospace' }}>{key}</Typography>
                      </Box>

                      <Box sx={{ flex: '1 1 auto', width: isMobile ? '100%' : undefined }}>
                        <TextField
                          fullWidth
                          value={value}
                          onChange={(e) => handleChange(key, e.target.value)}
                          size="small"
                        />
                      </Box>
                    </ListItem>
                  ))}
                </List>
              </Paper>
            )}
          </Grid>
        </Grid>
      </Box>

      <Snackbar open={saved} autoHideDuration={2000} onClose={() => setSaved(false)} anchorOrigin={{ vertical: 'bottom', horizontal: 'center' }}>
        <Alert severity="success" sx={{ width: '100%' }}>Properties updated successfully</Alert>
      </Snackbar>
    </PageContainer>
  );
}
