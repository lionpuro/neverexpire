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
		const str = new Date(d.textContent + " UTC").toLocaleDateString(
			window.navigator.language,
		);
		if (str !== "Invalid Date") {
			d.textContent = str;
		}
	});
}
localizeDates();
