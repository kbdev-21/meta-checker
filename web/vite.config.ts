import tailwindcss from '@tailwindcss/vite';
import adapter from '@sveltejs/adapter-static';
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [
		tailwindcss(),
		sveltekit({
			compilerOptions: {
				// Force runes mode for the project, except for libraries. Can be removed in svelte 6.
				runes: ({ filename }) => filename.split(/[/\\]/).includes('node_modules') ? undefined : true
			},
			// Route không prerender được (vd trang player) => build ra 200.html làm SPA fallback;
			// host tĩnh phải cấu hình trả file này cho URL không có file tương ứng.
			adapter: adapter({ fallback: '200.html' })
		})
	],
	// @lucide/svelte re-export file .svelte từ dist; không noExternal thì SSR để Node
	// import thẳng .svelte và chết ERR_UNKNOWN_FILE_EXTENSION.
	ssr: { noExternal: ['@lucide/svelte'] }
});
