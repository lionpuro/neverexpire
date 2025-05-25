import { join, resolve } from "path";
import { defineConfig } from "vite";
import { viteStaticCopy } from "vite-plugin-static-copy";

export default defineConfig({
	build: {
		emptyOutDir: true,
		outDir: join(__dirname, "assets/public/scripts"),
		rollupOptions: {
			external: [],
			input: [
				resolve(__dirname, "assets/src/local-time.ts"),
				resolve(__dirname, "assets/src/account.ts"),
			],
			output: {
				entryFileNames: "[name].min.js",
			},
		},
	},
	plugins: [
		viteStaticCopy({
			targets: [
				{
					src: resolve(__dirname, "node_modules/htmx.org/dist/htmx.min.js"),
					dest: resolve(__dirname, "assets/public/scripts"),
				},
			],
		}),
	],
});
