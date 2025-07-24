class LocalTime extends HTMLElement {
	public datetime?: string;
	public dateonly: boolean = false;
	public short: boolean = false;

	public constructor() {
		super();
	}

	public static get observedAttributes(): string[] {
		return ["datetime", "dateonly", "short"];
	}

	public attributeChangedCallback(
		name: string,
		_oldValue: string,
		newValue: string,
	) {
		switch (name) {
			case "datetime":
				this.datetime = newValue;
				break;
			case "dateonly":
				this.dateonly = newValue === "true";
				break;
			case "short":
				this.short = newValue === "true";
				break;
		}
	}

	connectedCallback() {
		if (!this.datetime) {
			console.warn("missing attribute datetime");
			return;
		}
		const date = new Date(this.datetime);
		if (isNaN(date.getTime())) {
			if (this.datetime !== "n/a") {
				console.error("local-time: invalid datetime input");
			}
			return;
		}
		let timeonly = false;
		if (this.short) {
			timeonly = true;
		}
		const datestr = localeString(date, this.dateonly, timeonly);
		if (datestr === "Invalid Date") {
			console.error("local-time: invalid datetime input");
			return;
		}
		this.textContent = datestr;
	}
}

function localeString(
	date: Date,
	dateonly: boolean,
	timeonly: boolean,
): string {
	if (timeonly) {
		const today = new Date().toLocaleDateString(window.navigator.language);
		const ts = date.toLocaleTimeString(window.navigator.language, {
			hour: "2-digit",
			minute: "2-digit",
		});
		const ds = date.toLocaleDateString(window.navigator.language);
		if (ds === today) {
			return ts;
		}
		return ds;
	}
	if (dateonly) {
		return date.toLocaleDateString(window.navigator.language);
	}
	return date.toLocaleString(window.navigator.language);
}

if (!customElements.get("local-time")) {
	customElements.define("local-time", LocalTime);
}
