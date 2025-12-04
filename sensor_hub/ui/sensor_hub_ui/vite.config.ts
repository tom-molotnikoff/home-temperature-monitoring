import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig(() => {
  const usePolling = process.env.HMR_POLLING === 'true'
  const pollingInterval = process.env.HMR_POLLING_INTERVAL ? Number(process.env.HMR_POLLING_INTERVAL) : 100

  return {
    plugins: [react()],
    server: {
      host: true,
      watch: usePolling ? { usePolling: true, interval: pollingInterval } : undefined,
    },
  }
})
