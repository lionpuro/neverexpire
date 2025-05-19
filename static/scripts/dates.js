function localizeDates() {
	const datetimes = document.querySelectorAll(".datetime");
	datetimes.forEach((d) => {
		const str = new Date(d.textContent + " UTC").toLocaleString(
			window.navigator.language,
		);
		if (str !== "Invalid Date") {
			d.textContent = str;
		}
	});
	const dates = document.querySelectorAll(".date");
	dates.forEach((d) => {
		const parts = d.textContent.split("-");
		const date = new Date(Date.UTC(parts[0], parts[1] - 1, parts[2], 1, 0, 0));
		const str = date.toLocaleDateString(window.navigator.language);
		if (str !== "Invalid Date") {
			d.textContent = str;
		}
	});
}
localizeDates();
