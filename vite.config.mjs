import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
  root: 'web',
  publicDir: 'static',
  envPrefix: ['VITE_', 'APP_'],
  build: {
    outDir: 'public',
    emptyOutDir: true,
    assetsDir: 'assets',
  },
  plugins: [react()],
});
