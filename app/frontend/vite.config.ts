import {defineConfig} from 'vite';
import {svelte} from '@sveltejs/vite-plugin-svelte';
export default defineConfig({plugins:[svelte()],build:{outDir:'../../cmd/desktop/frontend/dist',emptyOutDir:true}});

