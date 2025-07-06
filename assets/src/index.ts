document.body.addEventListener("click", (e) => {
	const menu = document.querySelector<HTMLDetailsElement>("#account-menu");
	const toggle = document.querySelector<HTMLElement>("#menu-toggle");
	if (!menu || !toggle) {
		return;
	}
	if (!menu.open) return;
	const rect = menu.getBoundingClientRect();
	if (
		e.clientX < rect.left ||
		e.clientX > rect.right ||
		e.clientY < rect.top ||
		e.clientY > rect.bottom
	) {
		toggle.click();
	}
});
