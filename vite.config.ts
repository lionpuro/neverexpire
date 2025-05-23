import { join, resolve } from "path";
import { defineConfig } from "vite";

export default defineConfig({
	build: {
		emptyOutDir: false,
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
});
