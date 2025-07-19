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
				resolve(__dirname, "assets/src/index.ts"),
				resolve(__dirname, "assets/src/local-time.ts"),
				resolve(__dirname, "assets/src/account.ts"),
				resolve(__dirname, "assets/src/swagger-ui.ts"),
			],
			output: {
				entryFileNames: "[name].js",
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
				{
					src: resolve(__dirname, "assets/static/*"),
					dest: resolve(__dirname, "assets/public"),
				},
				{
					src: resolve(
						__dirname,
						"node_modules/swagger-ui/dist/swagger-ui.css",
					),
					dest: resolve(__dirname, "assets/public/css"),
				},
			],
		}),
	],
});
