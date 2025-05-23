(() => {
	const dialog = document.querySelector<HTMLDialogElement>("#confirm-dialog");
	const deleteBtn = document.querySelector<HTMLButtonElement>("#delete-btn");
	const cancelBtn = document.querySelector<HTMLButtonElement>("#cancel-btn");
	if (!dialog || !deleteBtn || !cancelBtn) {
		return;
	}
	dialog.addEventListener("click", (e) => {
		const dialogRect = dialog.getBoundingClientRect();
		const outside =
			e.clientX < dialogRect.left ||
			e.clientX > dialogRect.right ||
			e.clientY < dialogRect.top ||
			e.clientY > dialogRect.bottom;
		if (outside) {
			dialog.close();
		}
	});
	deleteBtn.addEventListener("click", () => {
		dialog.showModal();
	});
	cancelBtn.addEventListener("click", () => {
		dialog.close();
	});
})();
